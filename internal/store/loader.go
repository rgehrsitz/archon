package store

// Loader reads/writes project.json and nodes/<id>.json
// This is a placeholder; real IO and validation will be added.

type Loader struct{}

func NewLoader() *Loader { return &Loader{} }
