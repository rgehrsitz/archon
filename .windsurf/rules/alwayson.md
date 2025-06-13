---
trigger: always_on
---

Follow Go 1.22 conventions: run goimports -local github.com/archon—no other formatters.

Follow ESLint defaults plus the project .eslintrc.cjs rules (CI fails on warnings).

Per‑file test coverage ≥ 90 % after each commit touching that file.

Never modify a file that is not listed in the active Windsurf task.

If you change any public API, update docs/agent/API_SPECIFICATIONS.md in the same PR.

Ask for a 👍 from a maintainer before executing more than one step from Planning Mode.