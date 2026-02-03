package modals

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/search"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	uisearch "github.com/abdul-hamid-achik/gpeek/internal/ui/search"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type filePosition struct {
	startLine int // Line number where file header starts
	endLine   int // Line number where file content ends
	expanded  bool
}

type DiffModal struct {
	BaseModal
	styles      *ui.Styles
	viewport    viewport.Model
	commitInfo  string       // Commit hash + message for title
	rawDiff     string       // Keep for reference
	parsedDiff  *diff.Diff   // Parsed once at creation
	expanded    map[int]bool // File index → expanded state
	allExpanded bool         // Toggle all state
	focusedFile int          // Currently focused file index

	// File position tracking for navigation
	filePositions []filePosition

	// Diff search
	diffSearch *uisearch.DiffSearch
}

func NewDiffModal(styles *ui.Styles, title, diffContent string, width, height int) *DiffModal {
	vp := viewport.New(width-4, height-6)

	// Parse diff once
	parsedDiff := diff.Parse(diffContent)

	// Initialize all files as collapsed
	expanded := make(map[int]bool)
	for i := range parsedDiff.Files {
		expanded[i] = false
	}

	m := &DiffModal{
		styles:        styles,
		viewport:      vp,
		commitInfo:    title,
		rawDiff:       diffContent,
		parsedDiff:    parsedDiff,
		expanded:      expanded,
		allExpanded:   false,
		focusedFile:   0,
		filePositions: make([]filePosition, len(parsedDiff.Files)),
		diffSearch:    uisearch.NewDiffSearch(styles),
	}
	m.width = width
	m.height = height
	m.diffSearch.SetWidth(width - 4)

	m.renderContent()
	return m
}

// fileLine represents a line in the file content for constrained rendering
type fileLine struct {
	isHunk bool
	text   string
	line   diff.Line
}

func (m *DiffModal) renderContent() {
	if len(m.parsedDiff.Files) == 0 {
		m.viewport.SetContent("No changes in this commit")
		return
	}

	var content strings.Builder
	lineNum := 0

	for i, file := range m.parsedDiff.Files {
		// Track file start position
		m.filePositions[i].startLine = lineNum
		m.filePositions[i].expanded = m.expanded[i]

		// Render file header with expand/collapse indicator
		indicator := "▶"
		if m.expanded[i] {
			indicator = "▼"
		}

		// Calculate stats for this file
		adds, dels := m.countFileChanges(file)

		// Determine filename to display
		filename := file.NewName
		if filename == "" || filename == "/dev/null" {
			filename = file.OldName
		}

		// Build stats string
		var stats string
		if file.IsBinary {
			stats = "(binary)"
		} else {
			stats = fmt.Sprintf("+%d -%d", adds, dels)
		}

		// Style header (highlight if focused)
		header := fmt.Sprintf("%s %s  (%s)", indicator, filename, stats)

		if i == m.focusedFile {
			// Focused file header style
			headerStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.styles.Theme.Background)).
				Background(lipgloss.Color(m.styles.Theme.Primary)).
				Bold(true)
			content.WriteString(headerStyle.Render(header))
		} else {
			// Normal file header style
			content.WriteString(m.styles.DiffMeta.Render(header))
		}
		content.WriteString("\n")
		lineNum++

		// Render file content if expanded (and not binary)
		if m.expanded[i] && !file.IsBinary {
			// Collect all lines for this file
			var fileLines []fileLine
			for _, hunk := range file.Hunks {
				fileLines = append(fileLines, fileLine{isHunk: true, text: hunk.Header})
				for _, line := range hunk.Lines {
					fileLines = append(fileLines, fileLine{line: line})
				}
			}

			// Render ALL lines (no constraint)
			for j := 0; j < len(fileLines); j++ {
				fl := fileLines[j]
				if fl.isHunk {
					content.WriteString(m.styles.DiffHunk.Render(fl.text))
				} else {
					content.WriteString(m.renderLine(fl.line, lineNum))
				}
				content.WriteString("\n")
				lineNum++
			}

			content.WriteString("\n")
			lineNum++
		}

		// Track file end position
		m.filePositions[i].endLine = lineNum
	}

	contentStr := content.String()
	m.viewport.SetContent(contentStr)
	m.diffSearch.SetContent(contentStr)
}

func (m *DiffModal) countFileChanges(file diff.FileDiff) (adds, dels int) {
	for _, hunk := range file.Hunks {
		for _, line := range hunk.Lines {
			switch line.Type {
			case diff.DiffAdd:
				adds++
			case diff.DiffRemove:
				dels++
			}
		}
	}
	return
}

func (m *DiffModal) renderLine(line diff.Line, lineNum int) string {
	prefix := " "
	var baseStyle lipgloss.Style

	switch line.Type {
	case diff.DiffAdd:
		prefix = "+"
		baseStyle = m.styles.DiffAdd
	case diff.DiffRemove:
		prefix = "-"
		baseStyle = m.styles.DiffRemove
	default:
		baseStyle = m.styles.DiffContext
	}

	content := line.Content

	// Apply search highlighting if matches exist
	matches := m.diffSearch.GetLineMatches(lineNum)
	if len(matches) > 0 {
		// Convert to search.Match format and highlight
		var searchMatches []search.Match
		for _, match := range matches {
			searchMatches = append(searchMatches, search.Match{
				Start: match.StartCol,
				End:   match.EndCol,
			})
		}
		h := search.NewHighlighter(m.styles.SearchMatch, baseStyle)
		return baseStyle.Render(prefix) + h.Highlight(content, searchMatches)
	}

	return baseStyle.Render(prefix + content)
}

func (m *DiffModal) isAllCollapsed() bool {
	for _, exp := range m.expanded {
		if exp {
			return false
		}
	}
	return true
}

// getVisibleFileIndex returns which file is currently most visible in the viewport
func (m *DiffModal) getVisibleFileIndex() int {
	viewMiddle := m.viewport.YOffset + m.viewport.Height/2

	for i, pos := range m.filePositions {
		if viewMiddle >= pos.startLine && viewMiddle < pos.endLine {
			return i
		}
	}

	// Default to last file if viewport is past all files
	return len(m.filePositions) - 1
}

// scrollToFile scrolls the viewport to show the file header
func (m *DiffModal) scrollToFile(fileIdx int) {
	if fileIdx < 0 || fileIdx >= len(m.filePositions) {
		return
	}

	// Scroll to the file header line
	m.viewport.SetYOffset(m.filePositions[fileIdx].startLine)
}

func (m *DiffModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	// Handle diff search if active
	if m.diffSearch.IsActive() {
		searchCmd, scrollTo := m.diffSearch.Update(msg)
		if scrollTo >= 0 {
			m.viewport.SetYOffset(scrollTo)
		}
		return m, searchCmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.diffSearch.HasSearch() {
				m.diffSearch.Deactivate()
				return m, nil
			}
			return nil, nil
		case "/":
			m.diffSearch.Activate()
			return m, nil
		case "n":
			if m.diffSearch.HasSearch() {
				scrollTo := m.diffSearch.NextMatch()
				if scrollTo >= 0 {
					m.viewport.SetYOffset(scrollTo)
				}
				return m, nil
			}
		case "N":
			if m.diffSearch.HasSearch() {
				scrollTo := m.diffSearch.PrevMatch()
				if scrollTo >= 0 {
					m.viewport.SetYOffset(scrollTo)
				}
				return m, nil
			}
		case "enter", " ":
			// Toggle focused file expansion
			if len(m.parsedDiff.Files) > 0 {
				m.expanded[m.focusedFile] = !m.expanded[m.focusedFile]
				// Update allExpanded state
				m.allExpanded = !m.isAllCollapsed()
				m.renderContent()
				// Scroll to keep file header visible
				m.scrollToFile(m.focusedFile)
			}
		case "a":
			// Toggle all files
			if len(m.parsedDiff.Files) > 0 {
				m.allExpanded = !m.allExpanded
				for i := range m.parsedDiff.Files {
					m.expanded[i] = m.allExpanded
				}
				m.renderContent()
				m.scrollToFile(m.focusedFile)
			}
		case "j", "down":
			// Hybrid navigation:
			// - If file is expanded and not at bottom of content, scroll down
			// - Otherwise, move to next file
			if len(m.parsedDiff.Files) > 0 {
				currentPos := m.filePositions[m.focusedFile]

				if m.expanded[m.focusedFile] {
					// Check if there's more content below in current file
					viewBottom := m.viewport.YOffset + m.viewport.Height
					if currentPos.endLine > viewBottom {
						// More content below, scroll down
						m.viewport.ScrollDown(1)
						// Update focused file based on what's now visible
						m.focusedFile = m.getVisibleFileIndex()
						m.renderContent()
						return m, nil
					}
				}

				// At bottom of file or file collapsed - move to next file
				if m.focusedFile < len(m.parsedDiff.Files)-1 {
					m.focusedFile++
					m.renderContent()
					m.scrollToFile(m.focusedFile)
				}
			}
		case "k", "up":
			// Hybrid navigation:
			// - If file is expanded and not at top of content, scroll up
			// - Otherwise, move to previous file
			if len(m.parsedDiff.Files) > 0 {
				currentPos := m.filePositions[m.focusedFile]

				if m.expanded[m.focusedFile] {
					// Check if we're not at the top of the file content
					if m.viewport.YOffset > currentPos.startLine {
						// Can scroll up within file
						m.viewport.ScrollUp(1)
						// Update focused file based on what's now visible
						m.focusedFile = m.getVisibleFileIndex()
						m.renderContent()
						return m, nil
					}
				}

				// At top of file or file collapsed - move to prev file
				if m.focusedFile > 0 {
					m.focusedFile--
					m.renderContent()
					m.scrollToFile(m.focusedFile)
				}
			}
		case "J":
			// Always move to next file (Shift+j)
			if len(m.parsedDiff.Files) > 0 && m.focusedFile < len(m.parsedDiff.Files)-1 {
				m.focusedFile++
				m.renderContent()
				m.scrollToFile(m.focusedFile)
			}
		case "K":
			// Always move to previous file (Shift+k)
			if len(m.parsedDiff.Files) > 0 && m.focusedFile > 0 {
				m.focusedFile--
				m.renderContent()
				m.scrollToFile(m.focusedFile)
			}
		case "}":
			// Jump to next file
			if len(m.parsedDiff.Files) > 0 && m.focusedFile < len(m.parsedDiff.Files)-1 {
				m.focusedFile++
				m.renderContent()
				m.scrollToFile(m.focusedFile)
			}
		case "{":
			// Jump to previous file
			if len(m.parsedDiff.Files) > 0 && m.focusedFile > 0 {
				m.focusedFile--
				m.renderContent()
				m.scrollToFile(m.focusedFile)
			}
		case "ctrl+d":
			m.viewport.HalfPageDown()
			m.focusedFile = m.getVisibleFileIndex()
			m.renderContent()
		case "ctrl+u":
			m.viewport.HalfPageUp()
			m.focusedFile = m.getVisibleFileIndex()
			m.renderContent()
		case "g":
			m.viewport.GotoTop()
			m.focusedFile = 0
			m.renderContent()
		case "G":
			m.viewport.GotoBottom()
			m.focusedFile = len(m.parsedDiff.Files) - 1
			m.renderContent()
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *DiffModal) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Primary)).
		Background(lipgloss.Color(m.styles.Theme.Background)).
		Bold(true).
		Padding(0, 1)

	title := titleStyle.Render(m.commitInfo)

	// File count info
	fileCount := len(m.parsedDiff.Files)
	expandedCount := 0
	for _, exp := range m.expanded {
		if exp {
			expandedCount++
		}
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Secondary)).
		Background(lipgloss.Color(m.styles.Theme.Background))

	header := headerStyle.Render(fmt.Sprintf("%d files (%d expanded)", fileCount, expandedCount))

	content := m.viewport.View()

	scrollPercent := int(m.viewport.ScrollPercent() * 100)
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Muted)).
		Background(lipgloss.Color(m.styles.Theme.Background))

	filledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Primary))

	trackStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Subtle))

	scrollbar := filledStyle.Render(strings.Repeat("█", scrollPercent/10)) +
		trackStyle.Render(strings.Repeat("░", 10-scrollPercent/10))

	footer := footerStyle.Render("j/k nav • J/K file • {/} jump • enter toggle • a all • / search • q close  ") + scrollbar

	// Add search bar if active
	var searchBar string
	if m.diffSearch.IsActive() || m.diffSearch.HasSearch() {
		searchBar = m.diffSearch.View()
	}

	var body string
	if searchBar != "" {
		body = lipgloss.JoinVertical(lipgloss.Left,
			title,
			header,
			"",
			content,
			"",
			searchBar,
			footer,
		)
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left,
			title,
			header,
			"",
			content,
			"",
			footer,
		)
	}

	return m.styles.Modal.
		Width(m.width).
		Height(m.height).
		Render(body)
}
