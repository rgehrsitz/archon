# CRUSH.md

Purpose: guidance for autonomous agents working in this repo (build, test, lint, style).

Commands
- Full backend test run: go test ./... 
- Run a single Go package tests: go test ./internal/<pkg> -v
- Run a single Go test by name: go test ./internal/<pkg> -run ^TestName$ -v
- With env for indexless tests: ARCHON_DISABLE_INDEX=1 go test ./internal/<pkg> -run ^TestName$ -v
- Frontend dev: (frontend) npm run dev
- Frontend tests: (frontend) npm test
- Frontend single test: (frontend) npx vitest -t "test name" or npm run test -- -t "test name"
- Build app: wails build
- Dev: wails dev

Style & conventions
- Formatting: run gofmt/goimports (go fmt) for Go; Prettier for frontend. Enforce with pre-commit hooks.
- Imports: group stdlib first, then third-party, then internal; use goimports to auto-sort for Go.
- Naming: Go uses mixedCaps (CamelCase) for functions/types, short receiver names (r, s, n). Tests: TestSomething_Scenario.
- Types: prefer concrete small structs; use interfaces for boundaries only. Keep exported types documented.
- Error handling: return wrapped errors with context (fmt.Errorf("...: %w", err)) or use errors package; never ignore returned errors.
- Logging: use zerolog in backend; include correlation IDs where available.
- Tests: table-driven tests preferred; keep tests hermetic and fast. Use ARCHON_DISABLE_INDEX=1 to skip heavy index work.
- Files & layout: follow existing internal/ and frontend/ structure. New backend packages go under internal/.
- Security: do not commit secrets. Use .env for placeholders when needed.
- Commits: do not auto-commit; ask user before staging or committing changes.

Notes
- No .cursor rules or .github/copilot-instructions.md detected; follow repository CLAUDE.md guidance for build and testing.
- If you add env-based requirements, create a .env file with placeholders and inform the user.

💘 Generated with Crush
