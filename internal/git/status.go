package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// String returns the string representation of FileStatus
func (s FileStatus) String() string {
	switch s {
	case StatusModified:
		return "modified"
	case StatusAdded:
		return "added"
	case StatusDeleted:
		return "deleted"
	case StatusRenamed:
		return "renamed"
	case StatusCopied:
		return "copied"
	case StatusUntracked:
		return "untracked"
	case StatusIgnored:
		return "ignored"
	case StatusUpdatedButUnmerged:
		return "conflict"
	default:
		return "unmodified"
	}
}

// MarshalJSON implements json.Marshaler for FileStatus
func (s FileStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

type FileEntry struct {
	Path   string     `json:"path"`
	Status FileStatus `json:"status"`
}

type Status struct {
	Staged    []FileEntry `json:"staged"`
	Unstaged  []FileEntry `json:"unstaged"`
	Untracked []string    `json:"untracked"`
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
	// Use git reset to unstage just the specified file, not the entire repo
	cmd := exec.Command("git", "reset", "HEAD", "--", path)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unstage failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func (r *Repository) Discard(path string) error {
	// Restore just the specified file from HEAD, not the entire repo
	cmd := exec.Command("git", "checkout", "HEAD", "--", path)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		// File may be untracked (not in HEAD) - remove it from the working tree
		absPath := filepath.Join(r.path, path)
		if removeErr := os.Remove(absPath); removeErr != nil {
			return fmt.Errorf("discard failed: %s", strings.TrimSpace(string(output)))
		}
	}
	return nil
}
