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
