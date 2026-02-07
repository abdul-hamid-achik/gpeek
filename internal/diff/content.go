package diff

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FilePosition tracks the line range of a file section in rendered content
type FilePosition struct {
	StartLine int
	EndLine   int
	Expanded  bool
}

// ContentLine represents a line in the rendered diff (hunk header or diff line)
type ContentLine struct {
	IsHunk bool
	Text   string
	Line   Line
}

// ContentStyles holds the styles needed for rendering diff content
type ContentStyles struct {
	DiffMeta    lipgloss.Style
	DiffHunk    lipgloss.Style
	DiffAdd     lipgloss.Style
	DiffRemove  lipgloss.Style
	DiffContext lipgloss.Style
	SearchMatch lipgloss.Style
	FocusedFile lipgloss.Style // Style for focused file header
}

// LineMatch represents a search match on a specific line
type LineMatch struct {
	StartCol int
	EndCol   int
}

// LineMatchProvider is a function that returns search matches for a given line number
type LineMatchProvider func(lineNum int) []LineMatch

// HighlightFunc applies search highlighting to content with matches
type HighlightFunc func(content string, matches []LineMatch, baseStyle lipgloss.Style) string

// RenderFileHeader renders a file header line with expand/collapse indicator and stats
func RenderFileHeader(file FileDiff, focused bool, expanded bool, styles ContentStyles) string {
	indicator := "▶"
	if expanded {
		indicator = "▼"
	}

	adds, dels := CountFileChanges(file)

	filename := file.NewName
	if filename == "" || filename == "/dev/null" {
		filename = file.OldName
	}

	var stats string
	if file.IsBinary {
		stats = "(binary)"
	} else {
		stats = fmt.Sprintf("+%d -%d", adds, dels)
	}

	header := fmt.Sprintf("%s %s  (%s)", indicator, filename, stats)

	if focused {
		return styles.FocusedFile.Render(header)
	}
	return styles.DiffMeta.Render(header)
}

// RenderDiffLine renders a single diff line with optional search highlighting
func RenderDiffLine(line Line, lineNum int, styles ContentStyles, matchProvider LineMatchProvider, highlightFn HighlightFunc) string {
	prefix := " "
	var baseStyle lipgloss.Style

	switch line.Type {
	case DiffAdd:
		prefix = "+"
		baseStyle = styles.DiffAdd
	case DiffRemove:
		prefix = "-"
		baseStyle = styles.DiffRemove
	default:
		baseStyle = styles.DiffContext
	}

	content := line.Content

	if matchProvider != nil && highlightFn != nil {
		matches := matchProvider(lineNum)
		if len(matches) > 0 {
			return baseStyle.Render(prefix) + highlightFn(content, matches, baseStyle)
		}
	}

	return baseStyle.Render(prefix + content)
}

// CountFileChanges counts additions and deletions in a file diff
func CountFileChanges(file FileDiff) (adds, dels int) {
	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			switch line.Type {
			case DiffAdd:
				adds++
			case DiffRemove:
				dels++
			}
		}
	}
	return
}

// IsAllCollapsed returns true if all files are collapsed
func IsAllCollapsed(expanded map[int]bool) bool {
	for _, exp := range expanded {
		if exp {
			return false
		}
	}
	return true
}

// RenderContent renders the full diff content with collapsible file sections.
// Returns the rendered string, updated file positions, and the final line count.
func RenderContent(
	parsed *Diff,
	expanded map[int]bool,
	focusedFile int,
	styles ContentStyles,
	matchProvider LineMatchProvider,
	highlightFn HighlightFunc,
) (string, []FilePosition) {
	if parsed == nil || len(parsed.Files) == 0 {
		return "No changes", nil
	}

	var content strings.Builder
	lineNum := 0
	positions := make([]FilePosition, len(parsed.Files))

	for i, file := range parsed.Files {
		positions[i].StartLine = lineNum
		positions[i].Expanded = expanded[i]

		// Render file header
		content.WriteString(RenderFileHeader(file, i == focusedFile, expanded[i], styles))
		content.WriteString("\n")
		lineNum++

		// Render file content if expanded (and not binary)
		if expanded[i] && !file.IsBinary {
			for _, hunk := range file.Hunks {
				content.WriteString(styles.DiffHunk.Render(hunk.Header))
				content.WriteString("\n")
				lineNum++

				for _, line := range hunk.Lines {
					content.WriteString(RenderDiffLine(line, lineNum, styles, matchProvider, highlightFn))
					content.WriteString("\n")
					lineNum++
				}
			}
			content.WriteString("\n")
			lineNum++
		}

		positions[i].EndLine = lineNum
	}

	return content.String(), positions
}
