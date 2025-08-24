package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	semdiff "github.com/rgehrsitz/archon/internal/diff/semantic"
	"github.com/rgehrsitz/archon/internal/git"
	"github.com/rgehrsitz/archon/internal/merge"
	"github.com/rgehrsitz/archon/internal/snapshot"
	"github.com/rgehrsitz/archon/internal/store"
)

func main() {
	var (
		projectPath string
	)
	flag.StringVar(&projectPath, "project", "", "Path to an Archon project (optional)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Archon CLI (MVP)\n")
		fmt.Fprintf(os.Stderr, "Usage: archon [--project path] <command> [args]\n")
		fmt.Fprintf(os.Stderr, "Commands: open | index | snapshot | diff | merge | attachment | export (stubs)\n")
		fmt.Fprintf(os.Stderr, "\nDiff usage:\n  archon --project <path> diff [--summary-only] [--json] [--semantic] <from> <to>\n")
		fmt.Fprintf(os.Stderr, "\nMerge usage:\n  archon --project <path> merge [--dry-run] [--json] <base> <ours> <theirs>\n")
		fmt.Fprintf(os.Stderr, "\nAttachment usage:\n  archon --project <path> attachment <add|list|get|remove|verify|gc> [args]\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	cmd := flag.Arg(0)
	switch cmd {
	case "snapshot":
		if err := runSnapshot(projectPath, flag.Args()[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "snapshot:", err)
			os.Exit(1)
		}
	case "diff":
		if err := runDiff(projectPath, flag.Args()[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "diff:", err)
			os.Exit(1)
		}
	case "merge":
		if err := runMerge(projectPath, flag.Args()[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "merge:", err)
			os.Exit(1)
		}
	case "attachment":
		if err := runAttachment(projectPath, flag.Args()[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "attachment:", err)
			os.Exit(1)
		}
	case "open", "index", "export":
		fmt.Println("(stub)", cmd, projectPath)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(2)
	}
}

func runSnapshot(projectPath string, args []string) error {
	if projectPath == "" {
		return fmt.Errorf("--project path is required")
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: archon --project <path> snapshot <create|list|get|restore> [args]")
	}

	sub := args[0]
	mgr, err := snapshot.NewManager(projectPath)
	if err != nil {
		return err
	}
	defer mgr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch sub {
	case "create":
		if len(args) < 3 { // name, message
			return fmt.Errorf("usage: archon --project <path> snapshot create <name> <message> [description]")
		}
		req := snapshot.CreateRequest{
			Name:        args[1],
			Message:     args[2],
			Description: strings.Join(args[3:], " "),
		}
		snap, e := mgr.Create(ctx, req)
		if e != nil {
			return e
		}
		fmt.Printf("Created snapshot %s (%s)\n", snap.Name, snap.Hash)
		return nil
	case "list":
		snaps, e := mgr.List(ctx)
		if e != nil {
			return e
		}
		for _, s := range snaps {
			fmt.Printf("%s\t%s\t%s\n", s.Name, s.Hash, s.Message)
		}
		return nil
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: archon --project <path> snapshot get <name>")
		}
		s, e := mgr.Get(ctx, args[1])
		if e != nil {
			return e
		}
		fmt.Printf("%s\t%s\t%s\n", s.Name, s.Hash, s.Message)
		return nil
	case "restore":
		if len(args) < 2 {
			return fmt.Errorf("usage: archon --project <path> snapshot restore <name>")
		}
		if e := mgr.Restore(ctx, args[1]); e != nil {
			return e
		}
		fmt.Println("Restored", args[1])
		return nil
	default:
		return fmt.Errorf("unknown snapshot subcommand: %s", sub)
	}
}

func runDiff(projectPath string, args []string) error {
	if projectPath == "" {
		return fmt.Errorf("--project path is required")
	}
	// Subcommand flags
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	summaryOnly := fs.Bool("summary-only", false, "Print only the summary line (no per-file changes)")
	jsonOut := fs.Bool("json", false, "Output machine-readable JSON (full diff unless --summary-only is set)")
	semantic := fs.Bool("semantic", false, "Compute semantic (node-aware) diff instead of textual file diff")
	only := fs.String("only", "", "Filter semantic changes to a comma-separated list: added,removed,renamed,moved,property,order,attachment (semantic mode)")
	nameOnly := fs.Bool("name-only", false, "Show only file names of changes (textual diff mode)")
	nameStatus := fs.Bool("name-status", false, "Show status letter and file names (textual diff mode)")
	exitCode := fs.Bool("exit-code", false, "Set exit code to 1 if there are changes, 0 otherwise (textual diff mode)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rem := fs.Args()
	if len(rem) < 2 {
		return fmt.Errorf("usage: archon --project <path> diff [--summary-only] [--json] <from> <to>")
	}
	from, to := rem[0], rem[1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Semantic mode: compute and print JSON or summary
	if *semantic {
		res, env := semdiff.Diff(projectPath, from, to)
		if env.Code != "" {
			return fmt.Errorf("failed to compute semantic diff: %s", env.Message)
		}

		// Optional filtering
		filtered := *res
		if *only != "" {
			allow := map[string]bool{}
			for _, part := range strings.Split(*only, ",") {
				k := strings.ToLower(strings.TrimSpace(part))
				if k != "" {
					allow[k] = true
				}
			}
			filtered.Changes = nil
			for _, c := range res.Changes {
				if semanticAllowed(allow, c.Type) {
					filtered.Changes = append(filtered.Changes, c)
				}
			}
			filtered.Summary = summarizeSemantic(filtered.Changes)
		}
		if *jsonOut {
			if *summaryOnly {
				out := struct {
					From    string          `json:"from"`
					To      string          `json:"to"`
					Summary semdiff.Summary `json:"summary"`
				}{From: filtered.From, To: filtered.To, Summary: filtered.Summary}
				return json.NewEncoder(os.Stdout).Encode(out)
			}
			return json.NewEncoder(os.Stdout).Encode(filtered)
		}
		// Text summary
		fmt.Printf("%s..%s semantic: %d changes (added:%d removed:%d renamed:%d moved:%d props:%d order:%d)\n",
			filtered.From, filtered.To, filtered.Summary.Total, filtered.Summary.Added, filtered.Summary.Removed, filtered.Summary.Renamed,
			filtered.Summary.Moved, filtered.Summary.PropertyChanged, filtered.Summary.OrderChanged)
		if *summaryOnly {
			return nil
		}
		// If not summary-only and not JSON, keep output minimal for now
		for _, c := range filtered.Changes {
			fmt.Printf("%s\t%s\n", c.Type, c.NodeID)
		}
		return nil
	}

	repo, err := git.NewRepository(git.RepositoryConfig{Path: projectPath})
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	defer repo.Close()

	d, env := repo.GetDiff(ctx, from, to)
	if env.Code != "" {
		return fmt.Errorf("failed to compute diff: %s", env.Message)
	}

	// JSON output mode
	if *jsonOut {
		if *summaryOnly {
			out := struct {
				From    string          `json:"from"`
				To      string          `json:"to"`
				Summary git.DiffSummary `json:"summary"`
			}{From: d.From, To: d.To, Summary: d.Summary}
			return json.NewEncoder(os.Stdout).Encode(out)
		}
		return json.NewEncoder(os.Stdout).Encode(d)
	}

	// Text output mode
	// Summary
	fmt.Printf("%s..%s: %d files changed, %d insertions(+), %d deletions(-)\n",
		d.From, d.To, d.Summary.FilesChanged, d.Summary.Additions, d.Summary.Deletions)
	if *summaryOnly {
		return nil
	}

	// Per-file details (sorted by path for deterministic output)
	sort.SliceStable(d.Files, func(i, j int) bool { return d.Files[i].Path < d.Files[j].Path })
	for _, f := range d.Files {
		var status string
		switch f.Status {
		case git.FileStatusAdded:
			status = "A"
		case git.FileStatusModified:
			status = "M"
		case git.FileStatusDeleted:
			status = "D"
		case git.FileStatusRenamed:
			status = "R"
		default:
			status = string(f.Status)
		}

		path := f.Path
		if f.Status == git.FileStatusRenamed && f.OldPath != "" {
			path = fmt.Sprintf("%s -> %s", f.OldPath, f.Path)
		}
		if *nameOnly {
			fmt.Println(path)
			continue
		}
		if *nameStatus {
			fmt.Printf("%s\t%s\n", status, path)
			continue
		}
		fmt.Printf("%s\t%s\t+%d -%d\n", status, path, f.Additions, f.Deletions)
	}
	if *exitCode {
		if d.Summary.FilesChanged > 0 {
			os.Exit(1)
		}
	}
	return nil
}

// semanticAllowed maps CLI filter terms to change types
func semanticAllowed(allow map[string]bool, t semdiff.ChangeType) bool {
	if len(allow) == 0 {
		return true
	}
	switch t {
	case semdiff.ChangeNodeAdded:
		return allow["added"] || allow["add"]
	case semdiff.ChangeNodeRemoved:
		return allow["removed"] || allow["remove"] || allow["deleted"] || allow["delete"]
	case semdiff.ChangeNodeRenamed:
		return allow["renamed"] || allow["rename"]
	case semdiff.ChangeNodeMoved:
		return allow["moved"] || allow["move"]
	case semdiff.ChangePropertyChanged:
		return allow["property"] || allow["properties"] || allow["prop"]
	case semdiff.ChangeOrderChanged:
		return allow["order"] || allow["reorder"]
	case semdiff.ChangeAttachmentChanged:
		return allow["attachment"] || allow["attachments"]
	default:
		return false
	}
}

// summarizeSemantic recomputes Summary for a filtered list
func summarizeSemantic(changes []semdiff.Change) semdiff.Summary {
	var s semdiff.Summary
	s.Total = len(changes)
	for _, c := range changes {
		switch c.Type {
		case semdiff.ChangeNodeAdded:
			s.Added++
		case semdiff.ChangeNodeRemoved:
			s.Removed++
		case semdiff.ChangeNodeRenamed:
			s.Renamed++
		case semdiff.ChangeNodeMoved:
			s.Moved++
		case semdiff.ChangePropertyChanged:
			s.PropertyChanged++
		case semdiff.ChangeOrderChanged:
			s.OrderChanged++
		case semdiff.ChangeAttachmentChanged:
			s.AttachmentChanged++
		}
	}
	return s
}

func runMerge(projectPath string, args []string) error {
	if projectPath == "" {
		return fmt.Errorf("--project path is required")
	}

	// Subcommand flags
	fs := flag.NewFlagSet("merge", flag.ContinueOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be merged without applying changes")
	jsonOut := fs.Bool("json", false, "Output machine-readable JSON")
	verbose := fs.Bool("verbose", false, "Show detailed information about changes")

	if err := fs.Parse(args); err != nil {
		return err
	}

	rem := fs.Args()
	if len(rem) < 3 {
		return fmt.Errorf("usage: archon --project <path> merge [--dry-run] [--json] <base> <ours> <theirs>")
	}

	base, ours, theirs := rem[0], rem[1], rem[2]

	// Perform the three-way merge
	res, err := merge.ThreeWay(projectPath, base, ours, theirs)
	if err != nil {
		return fmt.Errorf("three-way merge failed: %w", err)
	}

	// JSON output mode
	if *jsonOut {
		return json.NewEncoder(os.Stdout).Encode(res)
	}

	// Text output mode
	totalChanges := len(res.OursOnly) + len(res.TheirsOnly)
	conflicts := len(res.Conflicts)

	fmt.Printf("Three-way merge: %s <- %s + %s\n", base, ours, theirs)
	fmt.Printf("Non-conflicting changes: %d (ours: %d, theirs: %d)\n", 
		totalChanges, len(res.OursOnly), len(res.TheirsOnly))
	
	if conflicts > 0 {
		fmt.Printf("Conflicts detected: %d\n", conflicts)
		
		if *verbose {
			fmt.Println("\nConflicts:")
			for _, conflict := range res.Conflicts {
				fmt.Printf("  - %s: %s (%s)\n", conflict.NodeID, conflict.Field, conflict.Rule)
			}
		}
		
		fmt.Println("\nResolve conflicts manually before applying changes.")
		return fmt.Errorf("merge has conflicts")
	}

	if totalChanges == 0 {
		fmt.Println("No changes to apply.")
		return nil
	}

	if *dryRun {
		fmt.Println("\nChanges to be applied (dry-run):")
		if *verbose {
			printChanges(res.OursOnly, "Ours")
			printChanges(res.TheirsOnly, "Theirs")
		}
		fmt.Printf("Use --dry-run=false to apply %d changes\n", totalChanges)
		return nil
	}

	// Apply the changes
	if err := res.Apply(projectPath); err != nil {
		return fmt.Errorf("failed to apply merge changes: %w", err)
	}

	fmt.Printf("Successfully applied %d changes\n", len(res.Applied))
	
	if *verbose {
		fmt.Println("\nApplied changes:")
		printChanges(res.Applied, "Applied")
	}

	return nil
}

func printChanges(changes []semdiff.Change, label string) {
	if len(changes) == 0 {
		return
	}
	
	fmt.Printf("\n%s changes:\n", label)
	for _, change := range changes {
		switch change.Type {
		case semdiff.ChangeNodeRenamed:
			fmt.Printf("  - Renamed %s: %s -> %s\n", change.NodeID, change.NameFrom, change.NameTo)
		case semdiff.ChangeNodeMoved:
			fmt.Printf("  - Moved %s: %s -> %s\n", change.NodeID, change.ParentFrom, change.ParentTo)
		case semdiff.ChangePropertyChanged:
			fmt.Printf("  - Properties changed %s: %d properties\n", change.NodeID, len(change.ChangedProperties))
		case semdiff.ChangeOrderChanged:
			fmt.Printf("  - Reordered children of %s: %d -> %d children\n", change.ParentID, len(change.OrderFrom), len(change.OrderTo))
		case semdiff.ChangeNodeAdded:
			fmt.Printf("  - Added node %s\n", change.NodeID)
		case semdiff.ChangeNodeRemoved:
			fmt.Printf("  - Removed node %s\n", change.NodeID)
		default:
			fmt.Printf("  - %s %s\n", change.Type, change.NodeID)
		}
	}
}

func runAttachment(projectPath string, args []string) error {
	if projectPath == "" {
		return fmt.Errorf("--project path is required")
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: archon --project <path> attachment <add|list|get|remove|verify|gc> [args]")
	}

	sub := args[0]
	attachStore := store.NewAttachmentStore(projectPath)

	// Configure Git LFS if we're in a Git repository
	if repo, err := git.NewRepository(git.RepositoryConfig{Path: projectPath}); err == nil {
		attachStore = attachStore.WithGitRepository(repo)
		defer repo.Close()
	}

	switch sub {
	case "add":
		return runAttachmentAdd(attachStore, args[1:])
	case "list":
		return runAttachmentList(attachStore, args[1:])
	case "get":
		return runAttachmentGet(attachStore, args[1:])
	case "remove", "rm":
		return runAttachmentRemove(attachStore, args[1:])
	case "verify":
		return runAttachmentVerify(attachStore, args[1:])
	case "gc":
		return runAttachmentGC(attachStore, args[1:])
	default:
		return fmt.Errorf("unknown attachment subcommand: %s", sub)
	}
}

func runAttachmentAdd(attachStore *store.AttachmentStore, args []string) error {
	// Subcommand flags
	fs := flag.NewFlagSet("attachment add", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "Output machine-readable JSON")
	name := fs.String("name", "", "Override filename for the attachment")
	
	if err := fs.Parse(args); err != nil {
		return err
	}
	
	rem := fs.Args()
	if len(rem) == 0 {
		return fmt.Errorf("usage: archon attachment add [--json] [--name filename] <file-path|->\nUse '-' to read from stdin")
	}
	
	filePath := rem[0]
	var reader io.Reader
	var filename string
	
	if filePath == "-" {
		reader = os.Stdin
		filename = "stdin"
		if *name != "" {
			filename = *name
		}
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()
		reader = file
		filename = filePath
		if *name != "" {
			filename = *name
		}
	}
	
	attachment, err := attachStore.Store(reader, filename)
	if err != nil {
		return fmt.Errorf("failed to store attachment: %w", err)
	}
	
	if *jsonOut {
		return json.NewEncoder(os.Stdout).Encode(attachment)
	}
	
	fmt.Printf("Stored attachment: %s (%d bytes)\n", attachment.Hash, attachment.Size)
	fmt.Printf("Filename: %s\n", attachment.Filename)
	return nil
}

func runAttachmentList(attachStore *store.AttachmentStore, args []string) error {
	// Subcommand flags
	fs := flag.NewFlagSet("attachment list", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "Output machine-readable JSON")
	
	if err := fs.Parse(args); err != nil {
		return err
	}
	
	attachments, err := attachStore.List()
	if err != nil {
		return fmt.Errorf("failed to list attachments: %w", err)
	}
	
	if *jsonOut {
		return json.NewEncoder(os.Stdout).Encode(attachments)
	}
	
	if len(attachments) == 0 {
		fmt.Println("No attachments found")
		return nil
	}
	
	fmt.Printf("Found %d attachment(s):\n", len(attachments))
	fmt.Printf("%-64s %-10s %-5s %s\n", "HASH", "SIZE", "LFS", "STORED")
	fmt.Println(strings.Repeat("-", 90))
	
	for _, att := range attachments {
		lfsFlag := " "
		if att.IsLFS {
			lfsFlag = "L"
		}
		fmt.Printf("%-64s %-10d %-5s %s\n", att.Hash, att.Size, lfsFlag, att.StoredAt.Format("2006-01-02 15:04"))
	}
	return nil
}

func runAttachmentGet(attachStore *store.AttachmentStore, args []string) error {
	// Subcommand flags
	fs := flag.NewFlagSet("attachment get", flag.ContinueOnError)
	output := fs.String("output", "", "Output file path (default: stdout)")
	
	if err := fs.Parse(args); err != nil {
		return err
	}
	
	rem := fs.Args()
	if len(rem) == 0 {
		return fmt.Errorf("usage: archon attachment get [--output file] <hash>")
	}
	
	hash := rem[0]
	reader, err := attachStore.Retrieve(hash)
	if err != nil {
		return fmt.Errorf("failed to retrieve attachment: %w", err)
	}
	defer reader.Close()
	
	var writer io.Writer = os.Stdout
	var outputFile *os.File
	
	if *output != "" {
		outputFile, err = os.Create(*output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputFile.Close()
		writer = outputFile
	}
	
	written, err := io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("failed to copy attachment data: %w", err)
	}
	
	if *output != "" {
		fmt.Printf("Retrieved attachment %s to %s (%d bytes)\n", hash, *output, written)
	}
	
	return nil
}

func runAttachmentRemove(attachStore *store.AttachmentStore, args []string) error {
	// Subcommand flags
	fs := flag.NewFlagSet("attachment remove", flag.ContinueOnError)
	force := fs.Bool("force", false, "Skip confirmation prompts")
	
	if err := fs.Parse(args); err != nil {
		return err
	}
	
	rem := fs.Args()
	if len(rem) == 0 {
		return fmt.Errorf("usage: archon attachment remove [--force] <hash>")
	}
	
	hash := rem[0]
	
	// Get info first to show what we're deleting
	info, err := attachStore.GetInfo(hash)
	if err != nil {
		return fmt.Errorf("failed to get attachment info: %w", err)
	}
	
	if !*force {
		fmt.Printf("Delete attachment %s (%d bytes, stored %s)? [y/N]: ", 
			hash, info.Size, info.StoredAt.Format("2006-01-02 15:04"))
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}
	
	if err := attachStore.Delete(hash); err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}
	
	fmt.Printf("Deleted attachment: %s\n", hash)
	return nil
}

func runAttachmentVerify(attachStore *store.AttachmentStore, args []string) error {
	// Subcommand flags
	fs := flag.NewFlagSet("attachment verify", flag.ContinueOnError)
	all := fs.Bool("all", false, "Verify all attachments")
	
	if err := fs.Parse(args); err != nil {
		return err
	}
	
	rem := fs.Args()
	
	if *all {
		attachments, err := attachStore.List()
		if err != nil {
			return fmt.Errorf("failed to list attachments: %w", err)
		}
		
		var failed []string
		fmt.Printf("Verifying %d attachment(s)...\n", len(attachments))
		
		for _, att := range attachments {
			if err := attachStore.Verify(att.Hash); err != nil {
				fmt.Printf("FAIL %s: %v\n", att.Hash, err)
				failed = append(failed, att.Hash)
			} else {
				fmt.Printf("OK   %s\n", att.Hash)
			}
		}
		
		if len(failed) > 0 {
			return fmt.Errorf("verification failed for %d attachment(s)", len(failed))
		}
		
		fmt.Println("All attachments verified successfully")
		return nil
	}
	
	if len(rem) == 0 {
		return fmt.Errorf("usage: archon attachment verify [--all] [hash...]")
	}
	
	var failed []string
	for _, hash := range rem {
		if err := attachStore.Verify(hash); err != nil {
			fmt.Printf("FAIL %s: %v\n", hash, err)
			failed = append(failed, hash)
		} else {
			fmt.Printf("OK   %s\n", hash)
		}
	}
	
	if len(failed) > 0 {
		return fmt.Errorf("verification failed for %d attachment(s)", len(failed))
	}
	
	return nil
}

func runAttachmentGC(_ *store.AttachmentStore, args []string) error {
	// Subcommand flags
	fs := flag.NewFlagSet("attachment gc", flag.ContinueOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be deleted without actually deleting")
	
	if err := fs.Parse(args); err != nil {
		return err
	}
	
	// TODO: This is a placeholder implementation. The real implementation would:
	// 1. Scan all nodes to find referenced attachment hashes
	// 2. Compare with stored attachments
	// 3. Delete unreferenced attachments
	
	if *dryRun {
		fmt.Println("Garbage collection (dry-run): not yet implemented")
		fmt.Println("This would scan all nodes and remove unreferenced attachments")
	} else {
		fmt.Println("Garbage collection: not yet implemented")
		fmt.Println("Use --dry-run to preview what would be deleted")
	}
	
	return fmt.Errorf("garbage collection not yet implemented")
}
