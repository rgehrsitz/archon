# Repository Guidelines

## Project Structure & Module Organization
Archon couples a Go backend with a Svelte frontend delivered via Wails. Backend code lives in `internal/` (domain modules like `internal/index`, `internal/merge`) with command wiring in `cmd/archon`. UI components stay in `frontend/src` with generated bindings in `frontend/wailsjs`. Shared docs and design notes live under `docs/`, prebuilt artifacts land in `build/`, and example assets sit in `examples/`. Keep test fixtures near their modules; large binaries should use Git LFS-backed locations referenced by snapshots.

## Build, Test, and Development Commands
Use `wails dev` from the repo root for a live desktop shell. Ship-ready bundles come from `wails build`. Backend-only iterations can rely on `go run ./cmd/archon`. Inside `frontend/`, run `npm install` once, then `npm run dev` for isolated UI work and `npm run build` to mirror the production bundle. Use `npm run preview` when validating a static export prior to packaging.

## Coding Style & Naming Conventions
Go files are gofmt’d (tabs, camelCase identifiers, and receiver names aligned). Keep packages lowercase and group related logic inside domain folders under `internal/`. Frontend code follows Svelte + TypeScript defaults: two-space indentation, PascalCase component files (e.g., `HierarchyPanel.svelte`), and kebab-case routes. Tailwind utility strings should stay declarative; move shared patterns into `frontend/src/lib`. Run `npm run check` before submitting UI changes and prefer `const` with explicit types when exporting shared helpers.

## Testing Guidelines
Backend modules rely on the standard library runner: `go test ./...` covers units and integration helpers. Co-locate `_test.go` files with their subjects and favor table-driven tests for diff or merge logic. Frontend tests use Vitest and Testing Library via `npm run test`; add `:coverage` when validating impact on bundles. Keep deterministic fixtures in `frontend/src/mocks`, and only update Vitest snapshots when UI shifts are intentional. Name Go tests after the behavior (`TestIndex_RebuildCreatesMissingTables`) to surface coverage in trace logs.

## Commit & Pull Request Guidelines
Commits follow the Conventional Commit prefixes seen in history (`feat:`, `fix:`, `chore:`). Write imperative subjects under 72 characters and scope the message to one logical change; use follow-up commits for experiments. Pull requests must summarize intent, list touched packages (`internal/index`, `frontend/src/routes`), link GitHub issues when available, and attach screenshots or CLI output (e.g., `go test ./...`) for user-facing edits. Request review whenever modifying snapshot, diff, or merge subsystems.
