package merge

// Three-way merge placeholder per ADR-003.

type Resolution struct {
	Conflicts []Conflict
}

func ThreeWay(base, ours, theirs string) (*Resolution, error) {
	return &Resolution{Conflicts: []Conflict{}}, nil
}
