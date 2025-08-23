package git

import (
	"github.com/rgehrsitz/archon/internal/git/cli"
	"github.com/rgehrsitz/archon/internal/git/gogit"
)

// Conversion functions between main git types and CLI types
// These ensure type compatibility across the hybrid implementation

func convertCLIStatus(s *cli.Status) *Status {
	if s == nil {
		return nil
	}
	return &Status{
		Branch:          s.Branch,
		IsClean:         s.IsClean,
		AheadBy:         s.AheadBy,
		BehindBy:        s.BehindBy,
		StagedFiles:     s.StagedFiles,
		ModifiedFiles:   s.ModifiedFiles,
		UntrackedFiles:  s.UntrackedFiles,
		ConflictedFiles: s.ConflictedFiles,
	}
}

func convertCLICommits(commits []cli.Commit) []Commit {
	result := make([]Commit, len(commits))
	for i, c := range commits {
		result[i] = Commit{
			Hash:      c.Hash,
			ShortHash: c.ShortHash,
			Message:   c.Message,
			Author: Author{
				Name:  c.Author.Name,
				Email: c.Author.Email,
			},
		}
	}
	return result
}

func convertCLICommit(c *cli.Commit) *Commit {
	if c == nil {
		return nil
	}
	return &Commit{
		Hash:      c.Hash,
		ShortHash: c.ShortHash,
		Message:   c.Message,
		Author: Author{
			Name:  c.Author.Name,
			Email: c.Author.Email,
		},
	}
}

func convertCLITags(tags []cli.Tag) []Tag {
	result := make([]Tag, len(tags))
	for i, t := range tags {
		result[i] = Tag{
			Name:       t.Name,
			Hash:       t.Hash,
			Message:    t.Message,
			IsSnapshot: t.IsSnapshot,
		}
	}
	return result
}

func convertCLIDiff(d *cli.Diff) *Diff {
	if d == nil {
		return nil
	}
	
	files := make([]FileDiff, len(d.Files))
	for i, f := range d.Files {
		files[i] = FileDiff{
			Path:      f.Path,
			OldPath:   f.OldPath,
			Status:    FileStatus(f.Status),
			Additions: f.Additions,
			Deletions: f.Deletions,
		}
	}
	
	return &Diff{
		From:  d.From,
		To:    d.To,
		Files: files,
	}
}

func convertAuthorToCLI(a *Author) *cli.Author {
	if a == nil {
		return nil
	}
	return &cli.Author{
		Name:  a.Name,
		Email: a.Email,
	}
}

// Conversion functions for go-git types
func convertGoGitStatus(s *gogit.Status) *Status {
	if s == nil {
		return nil
	}
	return &Status{
		Branch:          s.Branch,
		IsClean:         s.IsClean,
		AheadBy:         s.AheadBy,
		BehindBy:        s.BehindBy,
		StagedFiles:     s.StagedFiles,
		ModifiedFiles:   s.ModifiedFiles,
		UntrackedFiles:  s.UntrackedFiles,
		ConflictedFiles: s.ConflictedFiles,
	}
}

func convertGoGitCommits(commits []gogit.Commit) []Commit {
	result := make([]Commit, len(commits))
	for i, c := range commits {
		result[i] = Commit{
			Hash:      c.Hash,
			ShortHash: c.ShortHash,
			Message:   c.Message,
			Author: Author{
				Name:  c.Author.Name,
				Email: c.Author.Email,
			},
		}
	}
	return result
}

func convertGoGitTags(tags []gogit.Tag) []Tag {
	result := make([]Tag, len(tags))
	for i, t := range tags {
		result[i] = Tag{
			Name:       t.Name,
			Hash:       t.Hash,
			Message:    t.Message,
			IsSnapshot: t.IsSnapshot,
		}
	}
	return result
}

func convertGoGitDiff(d *gogit.Diff) *Diff {
	if d == nil {
		return nil
	}
	
	files := make([]FileDiff, len(d.Files))
	for i, f := range d.Files {
		files[i] = FileDiff{
			Path:      f.Path,
			OldPath:   f.OldPath,
			Status:    FileStatus(f.Status),
			Additions: f.Additions,
			Deletions: f.Deletions,
		}
	}
	
	return &Diff{
		From:  d.From,
		To:    d.To,
		Files: files,
	}
}