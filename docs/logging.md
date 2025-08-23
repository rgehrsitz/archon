# Logging Guide

Archon uses a structured logging system built on `zerolog` with file rotation via `lumberjack`. Logs are JSON by default and written to `<project>/logs/archon.log` with rotation settings per ADR-006.

## Overview

- Structured JSON logs with timestamps and caller info
- Console output (dev) and rotating file output
- Configurable per environment in `<project>/.archon/logging.json`
- Frontend access via Wails `LoggingService`

## Quick Start (Go)

```go
import "github.com/rgehrsitz/archon/internal/logging"

func doWork() {
    logging.Log().Info().Str("op", "doWork").Msg("starting")
    // ...
    logging.InfoMsg("done")
}
```

## Convenience Helpers

- `logging.Log()` returns the global logger for fluent chaining.
  - Example: `logging.Log().Warn().Str("node_id", id).Msg("missing property")`
- Simple message helpers for quick, unstructured messages:
  - `logging.TraceMsg("...")`
  - `logging.DebugMsg("...")`
  - `logging.InfoMsg("...")`
  - `logging.WarnMsg("...")`
  - `logging.ErrorMsg("...")`
  - `logging.FatalMsg("...")`

These helpers avoid the previous awkward `logging.Info().Info()` pattern.

## Adding Context

```go
logging.Log().WithContext(map[string]any{
    "request_id": reqID,
    "operation":  "index_build",
}).Info().Msg("started")

logging.WithError(err).Error().Msg("failed to open file")
```

Other helpers:
- `logging.WithRequestID(id)`
- `logging.WithOperation(name)`
- `logging.WithError(err)`

## Configuration

Configuration is stored per project at `<project>/.archon/logging.json` and applied on project create/open.

Keys and common aliases accepted by `UpdateEnvironmentConfig`:
- `log_level`: one of `trace|debug|info|warn|error|fatal`
- `console`: bool
- `file`: bool
- `directory` or `log_directory`: path to log dir (relative paths resolve against project root)
- `max_size` or `max_size_mb`: int
- `max_backups`: int
- `max_age` or `max_age_days`: int
- `compress`: bool

Example update from API/frontend:

```ts
await LoggingService.UpdateLoggingConfig({
  log_level: "debug",
  console: true,
  file: true,
  directory: "logs",
  max_size_mb: 20,
  max_backups: 7,
  max_age_days: 14,
  compress: true,
});
```

## Recent Logs (Frontend)

`LoggingService.GetRecentLogs(limit: number)` returns the last N log entries in chronological order. It is rotation-aware and reads `archon.log*`, including compressed `.gz` files.

## Lifecycle

- On project create/open: `logging.InitializeFromEnvironment(projectPath)` is invoked.
- On app shutdown: `logging.Shutdown()` is called to clean up.

## Environment Detection

Logging environment is detected from `ARCHON_ENV`, `ENV`, or `NODE_ENV` (default `development`).
