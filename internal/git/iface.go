package git

// Interface for hybrid git layer (system git porcelain + go-git fast reads)

type Repo interface{
	Clone(url, path string) error
	Pull() error
	Push() error
	SetRemote(name, url string) error
	Status() (string, error)
}
