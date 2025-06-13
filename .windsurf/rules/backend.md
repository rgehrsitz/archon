---
trigger: glob
globs: src/**/*.go
---

Run go vet ./... locally; fix all issues before commit.

Run go test ./... locally; all tests must pass.

No CGO usage; pure Go only.