package migrate

// stepV1 initializes schema to version 1. This is a no-op for projects
// already at version >=1. It exists to ensure forward-only migration from
// legacy projects that may have schemaVersion 0 or missing.

type stepV1 struct{}

func (s *stepV1) Version() int { return 1 }
func (s *stepV1) Name() string { return "Initialize schema v1" }

func (s *stepV1) IsApplied(ctx *Context) (bool, error) {
	return ctx.Project.SchemaVersion >= 1, nil
}

func (s *stepV1) Apply(ctx *Context) error {
	p := ctx.Project
	if p.SchemaVersion >= 1 {
		return nil
	}
	p.SchemaVersion = 1
	return ctx.Loader.SaveProject(p)
}

func init() {
	Register(&stepV1{})
}
