package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rgehrsitz/archon/internal/git"
	"github.com/rgehrsitz/archon/internal/snapshot"
)

func main() {
	var (
		projectPath string
	)
	flag.StringVar(&projectPath, "project", "", "Path to an Archon project (optional)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Archon CLI (MVP)\n")
		fmt.Fprintf(os.Stderr, "Usage: archon [--project path] <command> [args]\n")
		fmt.Fprintf(os.Stderr, "Commands: open | index | snapshot | diff | export (stubs)\n")
		fmt.Fprintf(os.Stderr, "\nDiff usage:\n  archon --project <path> diff [--summary-only] [--json] <from> <to>\n")
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

	// Per-file details
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
		fmt.Printf("%s\t%s\t+%d -%d\n", status, path, f.Additions, f.Deletions)
	}
	return nil
}
