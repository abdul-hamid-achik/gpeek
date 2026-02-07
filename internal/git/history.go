package git

import (
	"errors"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// FileHistoryOptions configures file history retrieval
type FileHistoryOptions struct {
	Limit  int    // Maximum commits to return
	Offset int    // Skip this many commits (for pagination)
	Since  string // Only commits after this date
	Until  string // Only commits before this date
}

// FileHistory returns the commit history for a specific file
func (r *Repository) FileHistory(path string, opts FileHistoryOptions) ([]Commit, error) {
	var commits []Commit

	head, err := r.repo.Head()
	if err != nil {
		return commits, nil
	}

	logOpts := &gogit.LogOptions{
		From:     head.Hash(),
		Order:    gogit.LogOrderCommitterTime,
		FileName: &path,
	}

	iter, err := r.repo.Log(logOpts)
	if err != nil {
		return nil, err
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	skipped := 0
	count := 0

	err = iter.ForEach(func(c *object.Commit) error {
		// Handle pagination offset
		if skipped < opts.Offset {
			skipped++
			return nil
		}

		// Stop when we've collected enough
		if count >= limit {
			return errIterLimit
		}

		var parents []string
		for _, p := range c.ParentHashes {
			parents = append(parents, p.String())
		}

		commits = append(commits, Commit{
			Hash:    c.Hash.String(),
			Message: firstLine(c.Message),
			Author:  c.Author.Name,
			Email:   c.Author.Email,
			Time:    c.Author.When,
			IsMerge: len(c.ParentHashes) > 1,
			Parents: parents,
		})

		count++
		return nil
	})

	if err != nil && !errors.Is(err, errIterLimit) {
		return nil, err
	}

	return commits, nil
}

// FileHistoryWithStats returns file history with diff stats per commit
type CommitWithStats struct {
	Commit
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
}

// FileHistoryWithStats returns the commit history for a specific file with stats
func (r *Repository) FileHistoryWithStats(path string, opts FileHistoryOptions) ([]CommitWithStats, error) {
	commits, err := r.FileHistory(path, opts)
	if err != nil {
		return nil, err
	}

	// For now, return commits without stats (stats calculation is expensive)
	// In the future, we could add a flag to enable stats calculation
	result := make([]CommitWithStats, len(commits))
	for i, c := range commits {
		result[i] = CommitWithStats{
			Commit:    c,
			Additions: 0,
			Deletions: 0,
		}
	}

	return result, nil
}
