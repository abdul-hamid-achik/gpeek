package git

import (
	"errors"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
)

var ErrNotARepository = errors.New("not a git repository")

type Repository struct {
	repo     *gogit.Repository
	path     string
	name     string
	worktree *gogit.Worktree
}

func Open(path string) (*Repository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	repo, err := gogit.PlainOpenWithOptions(absPath, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		if errors.Is(err, gogit.ErrRepositoryNotExists) {
			return nil, ErrNotARepository
		}
		return nil, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	name := filepath.Base(absPath)
	if gitDir, err := findGitDir(absPath); err == nil {
		name = filepath.Base(filepath.Dir(gitDir))
	}

	return &Repository{
		repo:     repo,
		path:     absPath,
		name:     name,
		worktree: wt,
	}, nil
}

func findGitDir(path string) (string, error) {
	current := path
	for {
		gitPath := filepath.Join(current, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			if info.IsDir() {
				return gitPath, nil
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New("not in a git repository")
		}
		current = parent
	}
}

func (r *Repository) Name() string {
	return r.name
}

func (r *Repository) Path() string {
	return r.path
}

func (r *Repository) IsValid() bool {
	return r.repo != nil
}
