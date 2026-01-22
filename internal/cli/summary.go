package cli

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/spf13/cobra"
)

// SummaryResponse represents the complete repository snapshot
type SummaryResponse struct {
	Repository    RepositoryInfo    `json:"repository"`
	Status        SummaryStatus     `json:"status"`
	RecentCommits []CommitInfo      `json:"recent_commits"`
	Branches      SummaryBranches   `json:"branches"`
	Stashes       SummaryStashes    `json:"stashes"`
	Tags          SummaryTags       `json:"tags"`
}

type SummaryStatus struct {
	Staged         []FileInfo `json:"staged"`
	Unstaged       []FileInfo `json:"unstaged"`
	Untracked      []string   `json:"untracked"`
	StagedCount    int        `json:"staged_count"`
	UnstagedCount  int        `json:"unstaged_count"`
	UntrackedCount int        `json:"untracked_count"`
	IsClean        bool       `json:"is_clean"`
	HasConflicts   bool       `json:"has_conflicts"`
}

type SummaryBranches struct {
	Current string       `json:"current"`
	Local   []BranchInfo `json:"local"`
	Count   int          `json:"count"`
}

type BranchInfo struct {
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	IsCurrent bool   `json:"is_current"`
	Upstream  string `json:"upstream,omitempty"`
}

type SummaryStashes struct {
	Count   int         `json:"count"`
	Entries []StashInfo `json:"entries"`
}

type StashInfo struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
	Branch  string `json:"branch,omitempty"`
}

type SummaryTags struct {
	Count int       `json:"count"`
	Tags  []TagInfo `json:"tags"`
}

type TagInfo struct {
	Name        string `json:"name"`
	Hash        string `json:"hash"`
	IsAnnotated bool   `json:"is_annotated"`
}

var summaryCommitLimit int

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Complete repository snapshot",
	Long:  `Get a complete snapshot of the repository state in one call. Ideal for LLMs/agents that need full context.`,
	RunE:  runSummary,
}

func init() {
	summaryCmd.Flags().IntVarP(&summaryCommitLimit, "commits", "n", 10, "Number of recent commits to include")
}

func runSummary(cmd *cobra.Command, args []string) error {
	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	response, err := buildSummaryResponse(repo)
	if err != nil {
		return err
	}

	output(response, formatSummaryPlain)
	return nil
}

func buildSummaryResponse(repo *git.Repository) (SummaryResponse, error) {
	response := SummaryResponse{
		Repository: RepositoryInfo{
			Name:   repo.Name(),
			Path:   repo.Path(),
			Branch: repo.CurrentBranch(),
		},
	}

	// Get status
	status, err := repo.Status()
	if err == nil {
		staged := make([]FileInfo, len(status.Staged))
		for i, f := range status.Staged {
			staged[i] = FileInfo{Path: f.Path, Status: f.Status.String()}
		}
		unstaged := make([]FileInfo, len(status.Unstaged))
		for i, f := range status.Unstaged {
			unstaged[i] = FileInfo{Path: f.Path, Status: f.Status.String()}
		}

		hasConflicts := false
		for _, f := range status.Staged {
			if f.Status == git.StatusUpdatedButUnmerged {
				hasConflicts = true
				break
			}
		}

		response.Status = SummaryStatus{
			Staged:         staged,
			Unstaged:       unstaged,
			Untracked:      status.Untracked,
			StagedCount:    len(status.Staged),
			UnstagedCount:  len(status.Unstaged),
			UntrackedCount: len(status.Untracked),
			IsClean:        len(status.Staged) == 0 && len(status.Unstaged) == 0 && len(status.Untracked) == 0,
			HasConflicts:   hasConflicts,
		}
	}

	// Get recent commits
	commits, err := repo.Log(summaryCommitLimit)
	if err == nil {
		commitInfos := make([]CommitInfo, len(commits))
		for i, c := range commits {
			shortHash := c.Hash
			if len(shortHash) > 7 {
				shortHash = shortHash[:7]
			}
			commitInfos[i] = CommitInfo{
				Hash:      c.Hash,
				ShortHash: shortHash,
				Message:   c.Message,
				Author:    c.Author,
				Email:     c.Email,
				Time:      c.Time,
				TimeAgo:   timeAgo(c.Time),
				IsMerge:   c.IsMerge,
			}
		}
		response.RecentCommits = commitInfos
	}

	// Get branches
	branches, err := repo.ListBranches()
	if err == nil {
		branchInfos := make([]BranchInfo, len(branches))
		for i, b := range branches {
			branchInfos[i] = BranchInfo{
				Name:      b.Name,
				Hash:      b.Hash,
				IsCurrent: b.IsCurrent,
				Upstream:  b.Upstream,
			}
		}
		response.Branches = SummaryBranches{
			Current: repo.CurrentBranch(),
			Local:   branchInfos,
			Count:   len(branches),
		}
	}

	// Get stashes
	stashes, err := repo.StashList()
	if err == nil {
		stashInfos := make([]StashInfo, len(stashes))
		for i, s := range stashes {
			stashInfos[i] = StashInfo{
				Index:   s.Index,
				Message: s.Message,
				Branch:  s.Branch,
			}
		}
		response.Stashes = SummaryStashes{
			Count:   len(stashes),
			Entries: stashInfos,
		}
	}

	// Get tags
	tags, err := repo.ListTags()
	if err == nil {
		tagInfos := make([]TagInfo, len(tags))
		for i, t := range tags {
			tagInfos[i] = TagInfo{
				Name:        t.Name,
				Hash:        t.Hash,
				IsAnnotated: t.IsAnnotated,
			}
		}
		response.Tags = SummaryTags{
			Count: len(tags),
			Tags:  tagInfos,
		}
	}

	return response, nil
}

func formatSummaryPlain(data interface{}) string {
	response := data.(SummaryResponse)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Repository: %s\n", response.Repository.Name))
	sb.WriteString(fmt.Sprintf("Path: %s\n", response.Repository.Path))
	sb.WriteString(fmt.Sprintf("Branch: %s\n", response.Repository.Branch))
	sb.WriteString("\n")

	// Status
	sb.WriteString("=== Status ===\n")
	if response.Status.IsClean {
		sb.WriteString("Working tree clean\n")
	} else {
		if response.Status.StagedCount > 0 {
			sb.WriteString(fmt.Sprintf("Staged: %d files\n", response.Status.StagedCount))
		}
		if response.Status.UnstagedCount > 0 {
			sb.WriteString(fmt.Sprintf("Unstaged: %d files\n", response.Status.UnstagedCount))
		}
		if response.Status.UntrackedCount > 0 {
			sb.WriteString(fmt.Sprintf("Untracked: %d files\n", response.Status.UntrackedCount))
		}
	}
	sb.WriteString("\n")

	// Recent commits
	sb.WriteString("=== Recent Commits ===\n")
	for _, c := range response.RecentCommits {
		sb.WriteString(fmt.Sprintf("%s %s (%s)\n", c.ShortHash, c.Message, c.TimeAgo))
	}
	sb.WriteString("\n")

	// Branches
	sb.WriteString(fmt.Sprintf("=== Branches (%d) ===\n", response.Branches.Count))
	for _, b := range response.Branches.Local {
		marker := "  "
		if b.IsCurrent {
			marker = "* "
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", marker, b.Name))
	}
	sb.WriteString("\n")

	// Stashes
	if response.Stashes.Count > 0 {
		sb.WriteString(fmt.Sprintf("=== Stashes (%d) ===\n", response.Stashes.Count))
		for _, s := range response.Stashes.Entries {
			sb.WriteString(fmt.Sprintf("  stash@{%d}: %s\n", s.Index, s.Message))
		}
		sb.WriteString("\n")
	}

	// Tags
	if response.Tags.Count > 0 {
		sb.WriteString(fmt.Sprintf("=== Tags (%d) ===\n", response.Tags.Count))
		for _, t := range response.Tags.Tags {
			sb.WriteString(fmt.Sprintf("  %s\n", t.Name))
		}
	}

	return sb.String()
}
