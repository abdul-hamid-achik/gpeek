package cli

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/spf13/cobra"
)

// TagsResponse represents the JSON response for tags command
type TagsResponse struct {
	Tags  []TagDetailInfo `json:"tags"`
	Total int             `json:"total"`
}

type TagDetailInfo struct {
	Name        string `json:"name"`
	Hash        string `json:"hash"`
	ShortHash   string `json:"short_hash"`
	Message     string `json:"message,omitempty"`
	Tagger      string `json:"tagger,omitempty"`
	TimeAgo     string `json:"time_ago,omitempty"`
	IsAnnotated bool   `json:"is_annotated"`
}

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List tags",
	Long:  `Display all tags in the repository.`,
	RunE:  runTags,
}

func runTags(cmd *cobra.Command, args []string) error {
	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	tags, err := repo.ListTags()
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	response := buildTagsResponse(tags)
	output(response, formatTagsPlain)
	return nil
}

func buildTagsResponse(tags []git.Tag) TagsResponse {
	infos := make([]TagDetailInfo, len(tags))
	for i, t := range tags {
		shortHash := t.Hash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}
		info := TagDetailInfo{
			Name:        t.Name,
			Hash:        t.Hash,
			ShortHash:   shortHash,
			Message:     t.Message,
			Tagger:      t.Tagger,
			IsAnnotated: t.IsAnnotated,
		}
		if !t.TaggerTime.IsZero() {
			info.TimeAgo = timeAgo(t.TaggerTime)
		}
		infos[i] = info
	}
	return TagsResponse{
		Tags:  infos,
		Total: len(infos),
	}
}

func formatTagsPlain(data interface{}) string {
	response := data.(TagsResponse)
	var sb strings.Builder

	if response.Total == 0 {
		sb.WriteString("No tags\n")
		return sb.String()
	}

	sb.WriteString("Tags:\n")
	for _, t := range response.Tags {
		tagType := "lightweight"
		if t.IsAnnotated {
			tagType = "annotated"
		}
		sb.WriteString(fmt.Sprintf("  %s (%s) [%s]\n", t.Name, t.ShortHash, tagType))
		if t.Message != "" {
			// First line of message only
			msg := strings.Split(t.Message, "\n")[0]
			sb.WriteString(fmt.Sprintf("    %s\n", msg))
		}
	}

	sb.WriteString(fmt.Sprintf("\nTotal: %d tags\n", response.Total))
	return sb.String()
}
