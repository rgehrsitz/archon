package migrate

// Registry of idempotent forward-only migrations keyed by schemaVersion.

type Step func() error

var registry = map[int]Step{}

func Register(version int, step Step) { registry[version] = step }

func Run(current, target int) error {
	for v := current + 1; v <= target; v++ {
		if step, ok := registry[v]; ok {
			if err := step(); err != nil { return err }
		}
	}
	return nil
}
