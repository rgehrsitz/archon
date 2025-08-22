package store

// Project-related operations: create/open project structure and init LFS.

type ProjectStore struct{}

func NewProjectStore() *ProjectStore { return &ProjectStore{} }
