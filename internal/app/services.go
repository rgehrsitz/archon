package app

// Services aggregates constructed application services for binding.
// In MVP this is a placeholder for DI/wiring.
type Services struct{}

func NewServices() *Services { return &Services{} }
