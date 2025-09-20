# Archon

Archon is a desktop knowledge workbench for hierarchical projects with first-class snapshots, semantic diff/merge, and Git-backed sync. It uses a sharded on-disk JSON store (one file per node), content-addressed attachments via Git LFS, and a rebuildable SQLite index for fast search.

## Highlights

- **Storage**: `/project.json` + `/nodes/<id>.json` (authoritative), `/attachments/*` via LFS, rebuildable `/.archon/index/archon.db`.
- **Identity**: UUIDv7 IDs, sibling-level unique names, meaningful child order.
- **History**: Snapshots are commit + immutable tag; semantic diff/merge is the primary UX.
- **Plugins**: Comprehensive extensibility platform with 10 plugin types (Importers, Exporters, Transformers, Validators, Panels, Providers, etc.) running in sandboxed JS/TS workers with fine-grained permissions.
- **Stack**: Wails v2 + Go backend, Svelte 5 + Tailwind 4 + shadcn-svelte frontend.
- **Git**: Hybrid layer — system git for porcelain/LFS/creds; go-git for fast reads.
- **Reliability**: Error envelopes, rotating logs, autosave & crash safety.

Note for developers:

- Archon uses modernc.org/sqlite which guarantees FTS5 support across all platforms. Search indexing is always available.
- You can set `ARCHON_DISABLE_INDEX=1` to disable indexing in tests if needed.
- Snapshots are implemented as commit + immutable tag pairs; tags are created via the Git CLI and listed via go-git for speed.

## Quick Start

Prereqs: Go 1.23+, Node 18+, Wails CLI.

Dev:

```bash
wails dev
```

Build:

```bash
wails build
```

### CLI: Diff command

Compare two refs (commits or tags) and print a summary and per-file changes:

```bash
archon --project /path/to/project diff [--summary-only] [--json] [--semantic] [--only <filter>] <from> <to>
```

Flags:

- `--summary-only` prints just the one-line summary.
- `--json` emits machine-readable JSON (full diff by default; combine with `--summary-only` to emit only the summary object).
- `--semantic` performs semantic diff instead of textual diff, showing logical changes (renames, moves, property changes, etc.).
- `--only <filter>` (semantic only) filters to show only specific change types: `added`, `removed`, `renamed`, `moved`, `property`, `order`, `attachment`. Can specify multiple comma-separated values.

Examples:

```bash
# Human-readable summary + file list
archon --project ~/Projects/Example diff HEAD~1 HEAD

# Summary only
archon --project ~/Projects/Example diff --summary-only v1.0.0 v1.1.0

# JSON output
archon --project ~/Projects/Example diff --json snapshot-initial-state snapshot-updated-state

# Semantic diff showing all logical changes
archon --project ~/Projects/Example diff --semantic HEAD~1 HEAD

# Semantic diff showing only renames and moves
archon --project ~/Projects/Example diff --semantic --only renamed,moved HEAD~1 HEAD

# JSON semantic diff with summary only
archon --project ~/Projects/Example diff --semantic --json --summary-only v1.0.0 v1.1.0
```

### CLI: Merge command

Perform three-way semantic merges to combine changes from different branches or commits:

```bash
archon --project /path/to/project merge [--dry-run] [--json] [--verbose] <base> <ours> <theirs>
```

Flags:

- `--dry-run` shows what changes would be applied without modifying any files.
- `--json` emits machine-readable JSON output containing the merge resolution.
- `--verbose` displays detailed information about changes and conflicts.

The merge command performs intelligent semantic merging:

- **Non-conflicting changes** are automatically applied (renames, moves, property updates, reordering)
- **Conflicts** are detected when both sides modify the same logical field differently
- **Exit codes**: 0 for successful merge, 1 for conflicts detected

Examples:

```bash
# Basic three-way merge
archon --project ~/Projects/Example merge main-base feature-branch main-head

# Dry run to preview changes
archon --project ~/Projects/Example merge --dry-run base-commit our-commit their-commit

# JSON output for automation
archon --project ~/Projects/Example merge --json HEAD~2 HEAD~1 HEAD

# Verbose output showing detailed changes
archon --project ~/Projects/Example merge --verbose --dry-run base ours theirs
```

### CLI: Attachment command

Manage content-addressed file attachments with automatic deduplication and Git LFS integration:

```bash
archon --project /path/to/project attachment <add|list|get|remove|verify|gc> [args]
```

Subcommands:

- `add [--json] [--name filename] <file-path|->` stores a file as an attachment. Use `-` to read from stdin.
- `list [--json]` shows all stored attachments with hash, size, LFS status, and storage date.
- `get [--output file] <hash>` retrieves an attachment by hash. Outputs to stdout by default.
- `remove [--force] <hash>` deletes an attachment after confirmation (use `--force` to skip prompt).
- `verify [--all] [hash...]` verifies attachment integrity by recomputing SHA-256 hashes.
- `gc [--dry-run]` garbage collection for unreferenced attachments. Scans all project nodes and removes orphaned files.

The attachment system features:

- **Content-addressed storage** using SHA-256 hashing for deduplication
- **Automatic Git LFS integration** for files ≥1MB (configurable threshold)
- **Path sharding** for efficient file organization (`attachments/AB/ABC123...`)
- **Integrity verification** through hash validation
- **Reference validation** for attachment properties in nodes

Examples:

```bash
# Add a file attachment
archon --project ~/Projects/Example attachment add document.pdf

# Add from stdin with custom name
echo "Content" | archon --project ~/Projects/Example attachment add --name note.txt -

# List all attachments
archon --project ~/Projects/Example attachment list

# Get attachment content (outputs to terminal)
archon --project ~/Projects/Example attachment get a1b2c3d4...

# Save attachment to file
archon --project ~/Projects/Example attachment get --output retrieved.pdf a1b2c3d4...

# Verify all attachment integrity
archon --project ~/Projects/Example attachment verify --all

# Remove attachment (with confirmation)
archon --project ~/Projects/Example attachment remove a1b2c3d4...

# Force remove without confirmation
archon --project ~/Projects/Example attachment remove --force a1b2c3d4...

# Preview garbage collection (dry run)
archon --project ~/Projects/Example attachment gc --dry-run

# Run garbage collection to delete orphaned attachments
archon --project ~/Projects/Example attachment gc
```

## Project Layout (brief)

- `internal/` — core packages: `store/`, `index/sqlite/`, `git/`, `merge/`, `migrate/`, `api/`, `types/`.
- `frontend/` — Svelte 5 app (Tailwind 4 + shadcn-svelte), wrappers under `src/lib/api/`.
- `cmd/archon/` — CLI entry (stubs for automation flows).
- `build/` — platform assets/installers.

> Note: The section below documents the UI template used as a baseline for Archon.

## Tech Stack

This template combines the latest versions of powerful technologies:

- **[Wails v2.10.1](https://wails.io/)**: Build desktop applications using Go and web technologies
- **[Svelte v5.28.2](https://svelte.dev/)**: Cybernetically enhanced web apps with revolutionary reactivity
- **[Tailwind CSS v4.1.4](https://tailwindcss.com/)**: Utility-first CSS framework with new CSS-first configuration
- **[shadcn-svelte v1.0.6](https://shadcn-svelte.com/)**: Beautifully designed components built with Radix UI and Tailwind CSS
- **[TypeScript v5.8.3](https://www.typescriptlang.org/)**: JavaScript with syntax for types
- **[Vite v6.3.3](https://vitejs.dev/)**: Next generation frontend tooling for lightning-fast development

## Features

- **🎨 Complete UI Component Library**: 40+ pre-built, accessible components from shadcn-svelte
- **🌙 Dark Mode Support**: Built-in dark/light theme switching with proper color variables
- **⚡ Modern Development**: Svelte 5's runes system with Tailwind 4's CSS-first configuration
- **🔧 Type Safety**: Full TypeScript support throughout the project
- **🚀 Fast Development**: Hot module replacement powered by Vite with @tailwindcss/vite plugin
- **📱 Responsive Design**: Mobile-first approach with Tailwind's responsive utilities
- **♿ Accessibility**: Components built with accessibility best practices
- **🎯 Cross-Platform**: Build for Windows, macOS, and Linux with a single codebase
- **🔥 Go Backend**: Leverage Go's performance and ecosystem for your application logic

## UI Components Included

All shadcn-svelte components are pre-installed and ready to use:

- **Layout**: Card, Separator, Resizable, Sidebar
- **Navigation**: Breadcrumb, Menu, Navigation Menu, Pagination
- **Form**: Button, Input, Textarea, Select, Checkbox, Radio Group, Switch, Form
- **Data Display**: Table, Data Table, Avatar, Badge, Calendar, Chart
- **Feedback**: Alert, Alert Dialog, Dialog, Drawer, Sheet, Toast (Sonner), Progress, Skeleton
- **Overlay**: Popover, Tooltip, Hover Card, Context Menu, Dropdown Menu
- **And many more...**

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.23 or later)
- [Node.js](https://nodejs.org/) (version 16 or later)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### Template Quick Start

1. Clone this template:

```bash
git clone https://github.com/your-username/wails2-svelte5-tailwind4-ts-vite.git
cd wails2-svelte5-tailwind4-ts-vite
```

1. Install dependencies:

```bash
# Frontend dependencies are installed automatically by Wails
wails dev
```

### Development

To run in live development mode:

```bash
wails dev
```

This will:

- Start a Go backend server
- Launch a Vite development server with hot reload
- Open your application in a native window
- Enable access via browser at <http://localhost:34115>

For frontend-only development:

```bash
cd frontend
npm run dev
```

### Building

To build a production-ready distributable package:

```bash
wails build
```

## Using shadcn-svelte Components

Import and use components in your Svelte files:

```svelte
<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { Card, CardContent, CardHeader, CardTitle } from "$lib/components/ui/card";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
</script>

<Card class="w-96">
  <CardHeader>
    <CardTitle>Login</CardTitle>
  </CardHeader>
  <CardContent class="space-y-4">
    <div class="space-y-2">
      <Label for="email">Email</Label>
      <Input id="email" type="email" placeholder="Enter your email" />
    </div>
    <Button class="w-full">Sign In</Button>
  </CardContent>
</Card>
```

## Dark Mode

Dark mode is automatically configured. Toggle between themes:

```svelte
<script lang="ts">
  import { toggleMode } from "mode-watcher";
</script>

<Button on:click={toggleMode}>Toggle Theme</Button>
```

## Project Structure

```text
├── frontend/                    # Svelte frontend application
│   ├── src/
│   │   ├── lib/
│   │   │   ├── components/
│   │   │   │   └── ui/         # shadcn-svelte components
│   │   │   ├── utils.ts        # Utility functions
│   │   │   └── hooks/          # Custom Svelte hooks
│   │   ├── App.svelte          # Main application component
│   │   ├── main.ts             # Application entry point
│   │   └── style.css           # Global styles with Tailwind
│   ├── components.json         # shadcn-svelte configuration
│   ├── package.json            # Frontend dependencies
│   ├── tsconfig.json           # TypeScript configuration
│   └── vite.config.ts          # Vite configuration
├── app.go                      # Go application context
├── main.go                     # Go application entry point
├── wails.json                  # Wails configuration
└── build/                      # Build assets and configuration
```

## Configuration

### Tailwind CSS 4

This template uses Tailwind CSS 4 with the new CSS-first configuration approach. All theme variables are defined in `frontend/src/style.css`:

- CSS custom properties for colors
- Built-in dark mode support
- tw-animate-css for animations
- @tailwindcss/vite for optimal performance

### shadcn-svelte Components

Components are configured in `frontend/components.json` and installed in `frontend/src/lib/components/ui/`. Each component is:

- Fully customizable and owns its own code
- Built with accessibility in mind
- Styled with Tailwind CSS
- TypeScript ready

### Path Aliases

The following path aliases are configured:

- `$lib` → `frontend/src/lib`
- `$lib/components` → `frontend/src/lib/components`
- `$lib/components/ui` → `frontend/src/lib/components/ui`
- `$lib/utils` → `frontend/src/lib/utils`
- `$lib/hooks` → `frontend/src/lib/hooks`

## Adding New Components

To add additional shadcn-svelte components:

```bash
cd frontend
npx shadcn-svelte@latest add [component-name]
```

For example:

```bash
npx shadcn-svelte@latest add calendar
npx shadcn-svelte@latest add date-picker
```

## Customization

This template provides a solid foundation that's easy to extend:

### Adding Custom Styles

- Modify `frontend/src/style.css` for global styles
- Customize color schemes by updating CSS custom properties
- Add custom Tailwind utilities using the `@layer` directive

### Extending Components

- All shadcn-svelte components are in your codebase and fully customizable
- Create new components in `frontend/src/lib/components/`
- Follow the established patterns for consistency

### Go Backend Integration

- Add your application logic in Go files
- Use Wails context for frontend-backend communication
- Leverage Go's standard library and ecosystem

### Environment Configuration

- Configure different environments in `wails.json`
- Set up environment variables for different build targets
- Customize build flags and assets per platform

## Development Tips

### Hot Reload

- Changes to Svelte components reload instantly
- Go code changes trigger automatic recompilation
- CSS changes apply immediately with Vite HMR

### Debugging

- Use browser dev tools for frontend debugging
- Access frontend in browser mode: `http://localhost:34115`
- Use Go debugging tools for backend investigation

### Performance

- Vite handles optimal bundling and tree shaking
- Tailwind CSS purges unused styles automatically
- shadcn-svelte components are lightweight and performant

## Browser Compatibility

This template supports modern browsers with:

- ES2020+ features
- CSS custom properties
- CSS Grid and Flexbox
- Modern JavaScript APIs

For broader compatibility, configure Vite's build target in `frontend/vite.config.ts`.

## License

This template is available under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- [Wails](https://wails.io/) for making desktop app development with Go and web technologies seamless
- [shadcn-svelte](https://shadcn-svelte.com/) for providing beautiful, accessible components
- The Svelte, Tailwind CSS, TypeScript, and Vite communities for their excellent tools and documentation

## Support

If you find this template helpful, please consider:

- ⭐ Starring the repository
- 🐛 Reporting issues
- 💡 Suggesting improvements
- 📖 Contributing to documentation
