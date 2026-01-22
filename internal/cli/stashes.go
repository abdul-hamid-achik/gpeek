package cli

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/spf13/cobra"
)

// StashesResponse represents the JSON response for stashes command
type StashesResponse struct {
	Stashes []StashDetailInfo `json:"stashes"`
	Total   int               `json:"total"`
}

type StashDetailInfo struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
	Branch  string `json:"branch,omitempty"`
	Hash    string `json:"hash"`
	TimeAgo string `json:"time_ago"`
}

var stashesCmd = &cobra.Command{
	Use:   "stashes",
	Short: "List stashes",
	Long:  `Display all stashed changes.`,
	RunE:  runStashes,
}

func runStashes(cmd *cobra.Command, args []string) error {
	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	stashes, err := repo.StashList()
	if err != nil {
		return fmt.Errorf("failed to list stashes: %w", err)
	}

	response := buildStashesResponse(stashes)
	output(response, formatStashesPlain)
	return nil
}

func buildStashesResponse(stashes []git.Stash) StashesResponse {
	infos := make([]StashDetailInfo, len(stashes))
	for i, s := range stashes {
		infos[i] = StashDetailInfo{
			Index:   s.Index,
			Message: s.Message,
			Branch:  s.Branch,
			Hash:    s.Hash,
			TimeAgo: timeAgo(s.Time),
		}
	}
	return StashesResponse{
		Stashes: infos,
		Total:   len(infos),
	}
}

func formatStashesPlain(data interface{}) string {
	response := data.(StashesResponse)
	var sb strings.Builder

	if response.Total == 0 {
		sb.WriteString("No stashes\n")
		return sb.String()
	}

	sb.WriteString("Stashes:\n")
	for _, s := range response.Stashes {
		sb.WriteString(fmt.Sprintf("  stash@{%d}: %s (%s)\n", s.Index, s.Message, s.TimeAgo))
	}

	sb.WriteString(fmt.Sprintf("\nTotal: %d stashes\n", response.Total))
	return sb.String()
}
