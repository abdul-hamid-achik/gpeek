package cli

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/spf13/cobra"
)

// BranchesResponse represents the JSON response for branches command
type BranchesResponse struct {
	Current  string             `json:"current"`
	Branches []BranchDetailInfo `json:"branches"`
	Total    int                `json:"total"`
}

type BranchDetailInfo struct {
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	ShortHash string `json:"short_hash"`
	IsRemote  bool   `json:"is_remote"`
	IsCurrent bool   `json:"is_current"`
	Upstream  string `json:"upstream,omitempty"`
}

var branchesAll bool

var branchesCmd = &cobra.Command{
	Use:   "branches",
	Short: "List branches",
	Long:  `Display local and optionally remote branches.`,
	RunE:  runBranches,
}

func init() {
	branchesCmd.Flags().BoolVarP(&branchesAll, "all", "a", false, "Include remote branches")
}

func runBranches(cmd *cobra.Command, args []string) error {
	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	localBranches, err := repo.ListBranches()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	var allBranches []git.Branch
	allBranches = append(allBranches, localBranches...)

	if branchesAll {
		remoteBranches, err := repo.ListRemoteBranches()
		if err == nil {
			allBranches = append(allBranches, remoteBranches...)
		}
	}

	response := buildBranchesResponse(repo.CurrentBranch(), allBranches)
	output(response, formatBranchesPlain)
	return nil
}

func buildBranchesResponse(current string, branches []git.Branch) BranchesResponse {
	infos := make([]BranchDetailInfo, len(branches))
	for i, b := range branches {
		shortHash := b.Hash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}
		infos[i] = BranchDetailInfo{
			Name:      b.Name,
			Hash:      b.Hash,
			ShortHash: shortHash,
			IsRemote:  b.IsRemote,
			IsCurrent: b.IsCurrent,
			Upstream:  b.Upstream,
		}
	}
	return BranchesResponse{
		Current:  current,
		Branches: infos,
		Total:    len(infos),
	}
}

func formatBranchesPlain(data interface{}) string {
	response := data.(BranchesResponse)
	var sb strings.Builder

	sb.WriteString("Branches:\n")
	for _, b := range response.Branches {
		marker := "  "
		if b.IsCurrent {
			marker = "* "
		}
		name := b.Name
		if b.IsRemote {
			name = "remotes/" + name
		}
		sb.WriteString(fmt.Sprintf("%s%s (%s)\n", marker, name, b.ShortHash))
	}

	sb.WriteString(fmt.Sprintf("\nTotal: %d branches\n", response.Total))
	return sb.String()
}
