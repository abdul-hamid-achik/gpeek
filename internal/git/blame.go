package git

import (
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type BlameLine struct {
	LineNum int
	Hash    string
	Author  string
	Email   string
	Time    time.Time
	Content string
}

// Blame returns line-by-line blame information for a file
func (r *Repository) Blame(filepath string) ([]BlameLine, error) {
	// Get the HEAD commit
	head, err := r.repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	// Perform blame
	result, err := git.Blame(commit, filepath)
	if err != nil {
		return nil, err
	}

	var lines []BlameLine
	for i, line := range result.Lines {
		bl := BlameLine{
			LineNum: i + 1,
			Content: line.Text,
		}

		if line.Author != "" {
			bl.Author = line.Author
		}
		if line.Date != (time.Time{}) {
			bl.Time = line.Date
		}
		if line.Hash.String() != "" && line.Hash.String() != "0000000000000000000000000000000000000000" {
			bl.Hash = line.Hash.String()
		}

		lines = append(lines, bl)
	}

	return lines, nil
}

// BlameFile returns blame information with commit details looked up
func (r *Repository) BlameFile(filepath string) ([]BlameLine, error) {
	lines, err := r.Blame(filepath)
	if err != nil {
		return nil, err
	}

	// Cache for commit lookups to avoid repeated fetches
	commitCache := make(map[string]*object.Commit)

	for i := range lines {
		if lines[i].Hash == "" {
			continue
		}

		// Check cache first
		if cached, ok := commitCache[lines[i].Hash]; ok {
			lines[i].Author = cached.Author.Name
			lines[i].Email = cached.Author.Email
			lines[i].Time = cached.Author.When
			continue
		}

		// Look up commit for more details
		hash, err := r.repo.ResolveRevision(plumbing.Revision(lines[i].Hash))
		if err != nil {
			continue
		}

		commit, err := r.repo.CommitObject(*hash)
		if err != nil {
			continue
		}

		// Cache the commit
		commitCache[lines[i].Hash] = commit

		lines[i].Author = commit.Author.Name
		lines[i].Email = commit.Author.Email
		lines[i].Time = commit.Author.When
	}

	return lines, nil
}
