//go:build ignore

// This file is intentionally ignored by the Go build system.
// It was an early placeholder simple logger and is superseded by the
// structured zerolog-based logger defined in this package.
// Keeping the source for historical reference; not compiled.

package logging

import (
	"log"
	"os"
)

// Init sets up a simple rotating logger placeholder.
// TODO: Replace with zerolog rotating files (10MB x5) per spec.
func Init(logDir string) (*os.File, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(logDir+"/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	log.SetOutput(f)
	return f, nil
}
