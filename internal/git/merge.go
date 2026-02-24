package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Merge merges the given branch into the current branch.
func (r *Repository) Merge(branch string) error {
	cmd := exec.Command("git", "merge", branch)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("merge failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// Rebase rebases the current branch onto the given branch.
func (r *Repository) Rebase(onto string) error {
	cmd := exec.Command("git", "rebase", onto)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rebase failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// CherryPick applies the given commit to the current branch.
func (r *Repository) CherryPick(hash string) error {
	cmd := exec.Command("git", "cherry-pick", hash)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cherry-pick failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// ResetSoft moves HEAD to the given ref, keeping all changes staged.
func (r *Repository) ResetSoft(hash string) error {
	cmd := exec.Command("git", "reset", "--soft", hash)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reset failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// ResetHard moves HEAD to the given ref and discards all local changes.
// WARNING: This is a destructive operation and cannot be undone.
func (r *Repository) ResetHard(hash string) error {
	cmd := exec.Command("git", "reset", "--hard", hash)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reset failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// Revert creates a new commit that undoes the changes of the given commit.
func (r *Repository) Revert(hash string) error {
	cmd := exec.Command("git", "revert", "--no-edit", hash)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("revert failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}
