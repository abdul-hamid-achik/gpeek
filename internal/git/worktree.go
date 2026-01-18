package git

import (
	"os/exec"
	"strings"
)

type Worktree struct {
	Path   string
	Branch string
	Hash   string
	Bare   bool
}

func (r *Repository) ListWorktrees() ([]Worktree, error) {
	cmd := exec.Command("git", "-C", r.path, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseWorktreeList(string(output)), nil
}

func parseWorktreeList(output string) []Worktree {
	var worktrees []Worktree
	var current Worktree

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "HEAD ") {
			current.Hash = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") {
			branch := strings.TrimPrefix(line, "branch ")
			branch = strings.TrimPrefix(branch, "refs/heads/")
			current.Branch = branch
		} else if line == "bare" {
			current.Bare = true
		}
	}

	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees
}

func (r *Repository) AddWorktree(path, branch string) error {
	args := []string{"-C", r.path, "worktree", "add"}
	if branch != "" {
		args = append(args, "-b", branch)
	}
	args = append(args, path)
	if branch != "" {
		args = append(args, "HEAD")
	}

	cmd := exec.Command("git", args...)
	return cmd.Run()
}

func (r *Repository) RemoveWorktree(path string) error {
	cmd := exec.Command("git", "-C", r.path, "worktree", "remove", path)
	return cmd.Run()
}

func (r *Repository) PruneWorktrees() error {
	cmd := exec.Command("git", "-C", r.path, "worktree", "prune")
	return cmd.Run()
}
