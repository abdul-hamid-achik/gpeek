package git

import (
	gogit "github.com/go-git/go-git/v5"
)

type FileStatus int

const (
	StatusUnmodified FileStatus = iota
	StatusModified
	StatusAdded
	StatusDeleted
	StatusRenamed
	StatusCopied
	StatusUntracked
	StatusIgnored
	StatusUpdatedButUnmerged
)

type FileEntry struct {
	Path   string
	Status FileStatus
}

type Status struct {
	Staged    []FileEntry
	Unstaged  []FileEntry
	Untracked []string
}

func (r *Repository) Status() (*Status, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := wt.Status()
	if err != nil {
		return nil, err
	}

	result := &Status{}

	for path, s := range status {
		if s.Staging == gogit.Untracked && s.Worktree == gogit.Untracked {
			result.Untracked = append(result.Untracked, path)
			continue
		}

		if s.Staging != gogit.Unmodified && s.Staging != gogit.Untracked {
			result.Staged = append(result.Staged, FileEntry{
				Path:   path,
				Status: convertStatus(s.Staging),
			})
		}

		if s.Worktree != gogit.Unmodified && s.Worktree != gogit.Untracked {
			result.Unstaged = append(result.Unstaged, FileEntry{
				Path:   path,
				Status: convertStatus(s.Worktree),
			})
		}
	}

	return result, nil
}

func convertStatus(s gogit.StatusCode) FileStatus {
	switch s {
	case gogit.Modified:
		return StatusModified
	case gogit.Added:
		return StatusAdded
	case gogit.Deleted:
		return StatusDeleted
	case gogit.Renamed:
		return StatusRenamed
	case gogit.Copied:
		return StatusCopied
	case gogit.Untracked:
		return StatusUntracked
	case gogit.UpdatedButUnmerged:
		return StatusUpdatedButUnmerged
	default:
		return StatusUnmodified
	}
}

func (r *Repository) Stage(path string) error {
	wt, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	_, err = wt.Add(path)
	return err
}

func (r *Repository) StageAll() error {
	wt, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	err = wt.AddWithOptions(&gogit.AddOptions{All: true})
	return err
}

func (r *Repository) Unstage(path string) error {
	wt, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	_, err = wt.Remove(path)
	if err != nil {
		head, err := r.repo.Head()
		if err != nil {
			return err
		}

		commit, err := r.repo.CommitObject(head.Hash())
		if err != nil {
			return err
		}

		tree, err := commit.Tree()
		if err != nil {
			return err
		}

		_, err = tree.File(path)
		if err != nil {
			return wt.Reset(&gogit.ResetOptions{Mode: gogit.MixedReset})
		}
	}
	return err
}

func (r *Repository) Discard(path string) error {
	wt, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	head, err := r.repo.Head()
	if err != nil {
		return err
	}

	return wt.Reset(&gogit.ResetOptions{
		Commit: head.Hash(),
		Mode:   gogit.HardReset,
	})
}
