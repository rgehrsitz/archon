package gogit

// Placeholder for go-git fast read paths (log, tree walk, light diffs)

type Repo struct{}

func Open(path string) (*Repo, error) { return &Repo{}, nil }
