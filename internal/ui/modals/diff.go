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

	// Per-file scrolling
	fileScrollOffset map[int]int // File index → scroll offset within file
	maxLinesPerFile  int         // Max visible lines per expanded file

	// Diff search
	diffSearch *uisearch.DiffSearch
}

func NewDiffModal(styles *ui.Styles, title, diffContent string, width, height int) *DiffModal {
	vp := viewport.New(width-4, height-6)

	// Parse diff once
	parsedDiff := diff.Parse(diffContent)

	// Initialize all files as expanded
	expanded := make(map[int]bool)
	for i := range parsedDiff.Files {
		expanded[i] = true
	}

	m := &DiffModal{
		styles:           styles,
		viewport:         vp,
		commitInfo:       title,
		rawDiff:          diffContent,
		parsedDiff:       parsedDiff,
		expanded:         expanded,
		allExpanded:      true,
		focusedFile:      0,
		fileScrollOffset: make(map[int]int),
		maxLinesPerFile:  20,
		diffSearch:       uisearch.NewDiffSearch(styles),
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
		// Render file header with expand/collapse indicator
		indicator := "▶"
		if m.expanded[i] {
			indicator = "▼"
		}

		// Calculate stats for this file
		adds, dels := m.countFileChanges(file)
		totalLines := m.countFileLines(file)

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

		// Add scroll indicator to header if file has more content
		scrollInfo := ""
		if m.expanded[i] && totalLines > m.maxLinesPerFile {
			scrollInfo = fmt.Sprintf(" [%d/%d]", m.fileScrollOffset[i]+1, totalLines)
		}

		// Style header (highlight if focused)
		header := fmt.Sprintf("%s %s  (%s)%s", indicator, filename, stats, scrollInfo)

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

			offset := m.fileScrollOffset[i]

			// Show "↑ X more" if scrolled down
			if offset > 0 {
				content.WriteString(m.styles.Dim.Render(fmt.Sprintf("  ↑ %d more lines\n", offset)))
				lineNum++
			}

			// Render visible lines
			endLine := offset + m.maxLinesPerFile
			if endLine > len(fileLines) {
				endLine = len(fileLines)
			}

			for j := offset; j < endLine; j++ {
				fl := fileLines[j]
				if fl.isHunk {
					content.WriteString(m.styles.DiffHunk.Render(fl.text))
				} else {
					content.WriteString(m.renderLine(fl.line, lineNum))
				}
				content.WriteString("\n")
				lineNum++
			}

			// Show "↓ X more" if more content below
			remaining := len(fileLines) - endLine
			if remaining > 0 {
				content.WriteString(m.styles.Dim.Render(fmt.Sprintf("  ↓ %d more lines\n", remaining)))
				lineNum++
			}

			content.WriteString("\n")
			lineNum++
		}
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

func (m *DiffModal) countFileLines(file diff.FileDiff) int {
	count := 0
	for _, hunk := range file.Hunks {
		count++ // hunk header
		count += len(hunk.Lines)
	}
	return count
}

// scrollToFocusedFile scrolls the viewport to show the focused file header
func (m *DiffModal) scrollToFocusedFile() {
	if len(m.parsedDiff.Files) == 0 {
		return
	}

	// Calculate the line number where the focused file header appears
	lineNum := 0
	for i := 0; i < m.focusedFile; i++ {
		lineNum++ // File header line
		if m.expanded[i] && !m.parsedDiff.Files[i].IsBinary {
			totalLines := m.countFileLines(m.parsedDiff.Files[i])
			offset := m.fileScrollOffset[i]

			// "↑ X more" indicator
			if offset > 0 {
				lineNum++
			}

			// Visible lines
			endLine := offset + m.maxLinesPerFile
			if endLine > totalLines {
				endLine = totalLines
			}
			lineNum += endLine - offset

			// "↓ X more" indicator
			if endLine < totalLines {
				lineNum++
			}

			lineNum++ // Extra blank line after expanded file
		}
	}

	// Scroll to the calculated line
	m.viewport.SetYOffset(lineNum)
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
				// Reset scroll offset when collapsing
				if !m.expanded[m.focusedFile] {
					m.fileScrollOffset[m.focusedFile] = 0
				}
				// Update allExpanded state
				m.allExpanded = !m.isAllCollapsed()
				m.renderContent()
			}
		case "a":
			// Toggle all files
			if len(m.parsedDiff.Files) > 0 {
				m.allExpanded = !m.allExpanded
				for i := range m.parsedDiff.Files {
					m.expanded[i] = m.allExpanded
				}
				m.renderContent()
			}
		case "j", "down":
			if len(m.parsedDiff.Files) > 0 {
				file := m.parsedDiff.Files[m.focusedFile]
				totalLines := m.countFileLines(file)

				if m.expanded[m.focusedFile] && totalLines > m.maxLinesPerFile {
					// File is expanded and has scrollable content
					maxOffset := totalLines - m.maxLinesPerFile
					if m.fileScrollOffset[m.focusedFile] < maxOffset {
						// Scroll within file
						m.fileScrollOffset[m.focusedFile]++
						m.renderContent()
						return m, nil
					}
				}

				// At bottom of file or file collapsed - move to next file
				if m.focusedFile < len(m.parsedDiff.Files)-1 {
					m.focusedFile++
					m.fileScrollOffset[m.focusedFile] = 0 // Reset scroll for new file
					m.renderContent()
					m.scrollToFocusedFile()
				}
			}
		case "k", "up":
			if len(m.parsedDiff.Files) > 0 {
				if m.expanded[m.focusedFile] && m.fileScrollOffset[m.focusedFile] > 0 {
					// Scroll up within file
					m.fileScrollOffset[m.focusedFile]--
					m.renderContent()
					return m, nil
				}

				// At top of file or file collapsed - move to prev file
				if m.focusedFile > 0 {
					m.focusedFile--
					// Jump to bottom of previous file if expanded
					file := m.parsedDiff.Files[m.focusedFile]
					totalLines := m.countFileLines(file)
					if m.expanded[m.focusedFile] && totalLines > m.maxLinesPerFile {
						m.fileScrollOffset[m.focusedFile] = totalLines - m.maxLinesPerFile
					}
					m.renderContent()
					m.scrollToFocusedFile()
				}
			}
		case "ctrl+j", "J":
			// Scroll viewport content
			m.viewport.ScrollDown(1)
		case "ctrl+k", "K":
			// Scroll viewport content
			m.viewport.ScrollUp(1)
		case "ctrl+d":
			m.viewport.HalfPageDown()
		case "ctrl+u":
			m.viewport.HalfPageUp()
		case "g":
			m.viewport.GotoTop()
		case "G":
			m.viewport.GotoBottom()
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

	footer := footerStyle.Render("j/k nav • enter toggle • a toggle all • / search • q close  ") + scrollbar

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
