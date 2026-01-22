package git

import (
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Commit struct {
	Hash    string    `json:"hash"`
	Message string    `json:"message"`
	Author  string    `json:"author"`
	Email   string    `json:"email"`
	Time    time.Time `json:"time"`
	IsMerge bool      `json:"is_merge"`
	Parents []string  `json:"parents,omitempty"`
}

func (r *Repository) Commit(message string) error {
	wt, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	_, err = wt.Commit(message, &gogit.CommitOptions{})
	return err
}

func (r *Repository) AmendCommit(message string) error {
	wt, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	_, err = wt.Commit(message, &gogit.CommitOptions{
		Amend: true,
	})
	return err
}

func (r *Repository) GetLastCommitMessage() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", err
	}

	commit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return "", err
	}

	return commit.Message, nil
}

func (r *Repository) GetLastCommitInfo() (message, hash string, err error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", "", err
	}

	commit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return "", "", err
	}

	return commit.Message, head.Hash().String(), nil
}

func (r *Repository) Log(limit int) ([]Commit, error) {
	var commits []Commit

	head, err := r.repo.Head()
	if err != nil {
		return commits, nil
	}

	iter, err := r.repo.Log(&gogit.LogOptions{
		From:  head.Hash(),
		Order: gogit.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, err
	}

	count := 0
	err = iter.ForEach(func(c *object.Commit) error {
		if count >= limit {
			return nil
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

	if err != nil {
		return nil, err
	}

	return commits, nil
}

func (r *Repository) GetCommit(hash string) (*Commit, error) {
	h, err := r.repo.ResolveRevision(plumbing.Revision(hash))
	if err != nil {
		return nil, err
	}

	c, err := r.repo.CommitObject(*h)
	if err != nil {
		return nil, err
	}

	var parents []string
	for _, p := range c.ParentHashes {
		parents = append(parents, p.String())
	}

	return &Commit{
		Hash:    c.Hash.String(),
		Message: c.Message,
		Author:  c.Author.Name,
		Email:   c.Author.Email,
		Time:    c.Author.When,
		IsMerge: len(c.ParentHashes) > 1,
		Parents: parents,
	}, nil
}

func firstLine(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' {
			return s[:i]
		}
	}
	return s
}
