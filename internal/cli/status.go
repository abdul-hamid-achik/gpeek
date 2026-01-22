package cli

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/spf13/cobra"
)

// StatusResponse represents the JSON response for status command
type StatusResponse struct {
	Repository RepositoryInfo `json:"repository"`
	Staged     []FileInfo     `json:"staged"`
	Unstaged   []FileInfo     `json:"unstaged"`
	Untracked  []string       `json:"untracked"`
	Summary    StatusSummary  `json:"summary"`
}

type RepositoryInfo struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Branch string `json:"branch"`
}

type FileInfo struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

type StatusSummary struct {
	StagedCount    int  `json:"staged_count"`
	UnstagedCount  int  `json:"unstaged_count"`
	UntrackedCount int  `json:"untracked_count"`
	IsClean        bool `json:"is_clean"`
	HasConflicts   bool `json:"has_conflicts"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show repository status",
	Long:  `Display the current status of the repository including staged, unstaged, and untracked files.`,
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	status, err := repo.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	response := buildStatusResponse(repo, status)

	output(response, formatStatusPlain)
	return nil
}

func buildStatusResponse(repo *git.Repository, status *git.Status) StatusResponse {
	staged := make([]FileInfo, len(status.Staged))
	for i, f := range status.Staged {
		staged[i] = FileInfo{
			Path:   f.Path,
			Status: f.Status.String(),
		}
	}

	unstaged := make([]FileInfo, len(status.Unstaged))
	for i, f := range status.Unstaged {
		unstaged[i] = FileInfo{
			Path:   f.Path,
			Status: f.Status.String(),
		}
	}

	hasConflicts := false
	for _, f := range status.Staged {
		if f.Status == git.StatusUpdatedButUnmerged {
			hasConflicts = true
			break
		}
	}
	if !hasConflicts {
		for _, f := range status.Unstaged {
			if f.Status == git.StatusUpdatedButUnmerged {
				hasConflicts = true
				break
			}
		}
	}

	return StatusResponse{
		Repository: RepositoryInfo{
			Name:   repo.Name(),
			Path:   repo.Path(),
			Branch: repo.CurrentBranch(),
		},
		Staged:    staged,
		Unstaged:  unstaged,
		Untracked: status.Untracked,
		Summary: StatusSummary{
			StagedCount:    len(status.Staged),
			UnstagedCount:  len(status.Unstaged),
			UntrackedCount: len(status.Untracked),
			IsClean:        len(status.Staged) == 0 && len(status.Unstaged) == 0 && len(status.Untracked) == 0,
			HasConflicts:   hasConflicts,
		},
	}
}

func formatStatusPlain(data interface{}) string {
	response := data.(StatusResponse)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("On branch %s\n", response.Repository.Branch))

	if response.Summary.IsClean {
		sb.WriteString("nothing to commit, working tree clean\n")
		return sb.String()
	}

	if len(response.Staged) > 0 {
		sb.WriteString("\nChanges to be committed:\n")
		for _, f := range response.Staged {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", f.Status, f.Path))
		}
	}

	if len(response.Unstaged) > 0 {
		sb.WriteString("\nChanges not staged for commit:\n")
		for _, f := range response.Unstaged {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", f.Status, f.Path))
		}
	}

	if len(response.Untracked) > 0 {
		sb.WriteString("\nUntracked files:\n")
		for _, f := range response.Untracked {
			sb.WriteString(fmt.Sprintf("  %s\n", f))
		}
	}

	return sb.String()
}
