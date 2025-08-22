package merge

// Semantic diff placeholder per ADR-003.
// Detects rename/move/property/structure changes between refs.

type SemanticDiff struct {
	Changes []Change
}

func Diff(refA, refB string) (*SemanticDiff, error) {
	return &SemanticDiff{Changes: []Change{}}, nil
}
