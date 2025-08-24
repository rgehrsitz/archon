package logging

// NewTestLogger returns a quiet logger suitable for unit tests
func NewTestLogger() Logger {
	cfg := DefaultConfig()
	cfg.OutputConsole = false
	cfg.OutputFile = false
	l, _ := NewLogger(cfg)
	return *l
}
