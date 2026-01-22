package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/spf13/cobra"
)

// LogResponse represents the JSON response for log command
type LogResponse struct {
	Commits []CommitInfo `json:"commits"`
	Total   int          `json:"total"`
}

type CommitInfo struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"short_hash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Time      time.Time `json:"time"`
	TimeAgo   string    `json:"time_ago"`
	IsMerge   bool      `json:"is_merge"`
	Parents   []string  `json:"parents,omitempty"`
}

var (
	logLimit  int
	logAuthor string
	logSince  string
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show commit history",
	Long:  `Display commit history with optional filters.`,
	RunE:  runLog,
}

func init() {
	logCmd.Flags().IntVarP(&logLimit, "limit", "n", 50, "Number of commits to show")
	logCmd.Flags().StringVarP(&logAuthor, "author", "a", "", "Filter by author")
	logCmd.Flags().StringVar(&logSince, "since", "", "Show commits since date (e.g., '2024-01-01')")
}

func runLog(cmd *cobra.Command, args []string) error {
	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	commits, err := repo.Log(logLimit)
	if err != nil {
		return fmt.Errorf("failed to get log: %w", err)
	}

	// Apply filters
	filtered := filterCommits(commits, logAuthor, logSince)

	response := buildLogResponse(filtered)
	output(response, formatLogPlain)
	return nil
}

func filterCommits(commits []git.Commit, author, since string) []git.Commit {
	if author == "" && since == "" {
		return commits
	}

	var sinceTime time.Time
	if since != "" {
		var err error
		sinceTime, err = time.Parse("2006-01-02", since)
		if err != nil {
			// Try other formats
			sinceTime, _ = time.Parse(time.RFC3339, since)
		}
	}

	var filtered []git.Commit
	for _, c := range commits {
		if author != "" && !strings.Contains(strings.ToLower(c.Author), strings.ToLower(author)) {
			continue
		}
		if !sinceTime.IsZero() && c.Time.Before(sinceTime) {
			continue
		}
		filtered = append(filtered, c)
	}
	return filtered
}

func buildLogResponse(commits []git.Commit) LogResponse {
	infos := make([]CommitInfo, len(commits))
	for i, c := range commits {
		shortHash := c.Hash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}
		infos[i] = CommitInfo{
			Hash:      c.Hash,
			ShortHash: shortHash,
			Message:   c.Message,
			Author:    c.Author,
			Email:     c.Email,
			Time:      c.Time,
			TimeAgo:   timeAgo(c.Time),
			IsMerge:   c.IsMerge,
			Parents:   c.Parents,
		}
	}
	return LogResponse{
		Commits: infos,
		Total:   len(infos),
	}
}

func timeAgo(t time.Time) string {
	duration := time.Since(t)
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 30*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case duration < 365*24*time.Hour:
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

func formatLogPlain(data interface{}) string {
	response := data.(LogResponse)
	var sb strings.Builder

	for _, c := range response.Commits {
		sb.WriteString(fmt.Sprintf("%s %s <%s> %s\n", c.ShortHash, c.Author, c.Email, c.TimeAgo))
		sb.WriteString(fmt.Sprintf("    %s\n", c.Message))
		if c.IsMerge {
			sb.WriteString("    (merge commit)\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Total: %d commits\n", response.Total))
	return sb.String()
}
