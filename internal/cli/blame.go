package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/spf13/cobra"
)

// BlameResponse represents the JSON response for blame command
type BlameResponse struct {
	File  string          `json:"file"`
	Lines []BlameLineInfo `json:"lines"`
	Total int             `json:"total"`
}

type BlameLineInfo struct {
	LineNum int       `json:"line_num"`
	Hash    string    `json:"hash,omitempty"`
	Author  string    `json:"author,omitempty"`
	Email   string    `json:"email,omitempty"`
	Time    time.Time `json:"time,omitempty"`
	TimeAgo string    `json:"time_ago,omitempty"`
	Content string    `json:"content"`
}

var (
	blameStartLine int
	blameEndLine   int
)

var blameCmd = &cobra.Command{
	Use:   "blame <file>",
	Short: "Show line-by-line attribution",
	Long:  `Display who last modified each line of a file.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBlame,
}

func init() {
	blameCmd.Flags().IntVar(&blameStartLine, "start", 0, "Start line (0 for beginning)")
	blameCmd.Flags().IntVar(&blameEndLine, "end", 0, "End line (0 for end of file)")
}

func runBlame(cmd *cobra.Command, args []string) error {
	file := args[0]

	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	lines, err := repo.BlameFile(file)
	if err != nil {
		return fmt.Errorf("failed to get blame: %w", err)
	}

	// Apply line range filter
	if blameStartLine > 0 || blameEndLine > 0 {
		start := blameStartLine
		if start < 1 {
			start = 1
		}
		end := blameEndLine
		if end < 1 || end > len(lines) {
			end = len(lines)
		}
		if start <= end && start <= len(lines) {
			lines = lines[start-1 : end]
		}
	}

	response := buildBlameResponse(file, lines)
	output(response, formatBlamePlain)
	return nil
}

func buildBlameResponse(file string, lines []git.BlameLine) BlameResponse {
	infos := make([]BlameLineInfo, len(lines))
	for i, l := range lines {
		info := BlameLineInfo{
			LineNum: l.LineNum,
			Hash:    l.Hash,
			Author:  l.Author,
			Email:   l.Email,
			Time:    l.Time,
			Content: l.Content,
		}
		if !l.Time.IsZero() {
			info.TimeAgo = timeAgo(l.Time)
		}
		// Shorten hash for display
		if len(info.Hash) > 8 {
			info.Hash = info.Hash[:8]
		}
		infos[i] = info
	}
	return BlameResponse{
		File:  file,
		Lines: infos,
		Total: len(infos),
	}
}

func formatBlamePlain(data interface{}) string {
	response := data.(BlameResponse)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Blame for %s:\n\n", response.File))

	// Find max author name length for alignment
	maxAuthor := 0
	for _, l := range response.Lines {
		if len(l.Author) > maxAuthor {
			maxAuthor = len(l.Author)
		}
	}
	if maxAuthor > 20 {
		maxAuthor = 20
	}

	for _, l := range response.Lines {
		hash := l.Hash
		if hash == "" {
			hash = "        "
		}
		author := l.Author
		if len(author) > maxAuthor {
			author = author[:maxAuthor]
		}

		sb.WriteString(fmt.Sprintf("%s %-*s %4d | %s\n",
			hash, maxAuthor, author, l.LineNum, l.Content))
	}

	return sb.String()
}
