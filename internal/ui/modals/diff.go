package modals

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/search"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	uisearch "github.com/abdul-hamid-achik/gpeek/internal/ui/search"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

	// Side-by-side toggle
	sideBySide bool
	renderer   *diff.Renderer

	// File position tracking for navigation
	filePositions []diff.FilePosition

	// Diff search
	diffSearch *uisearch.DiffSearch
}

func NewDiffModal(styles *ui.Styles, title, diffContent string, width, height int) *DiffModal {
	vp := viewport.New()
	vp.SetWidth(width - 4)
	vp.SetHeight(height - 6)

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
		sideBySide:    false,
		renderer:      diff.NewRenderer(styles),
		filePositions: make([]diff.FilePosition, len(parsedDiff.Files)),
		diffSearch:    uisearch.NewDiffSearch(styles),
	}
	m.width = width
	m.height = height
	m.diffSearch.SetWidth(width - 4)

	m.renderContent()
	return m
}

func (m *DiffModal) contentStyles() diff.ContentStyles {
	return diff.ContentStyles{
		DiffMeta:    m.styles.DiffMeta,
		DiffHunk:    m.styles.DiffHunk,
		DiffAdd:     m.styles.DiffAdd,
		DiffRemove:  m.styles.DiffRemove,
		DiffContext:  m.styles.DiffContext,
		SearchMatch: m.styles.SearchMatch,
		FocusedFile: lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.styles.Theme.Background)).
			Background(lipgloss.Color(m.styles.Theme.Primary)).
			Bold(true),
		ContentWidth: m.viewport.Width(),
	}
}

func (m *DiffModal) matchProvider(lineNum int) []diff.LineMatch {
	matches := m.diffSearch.GetLineMatches(lineNum)
	if len(matches) == 0 {
		return nil
	}
	result := make([]diff.LineMatch, len(matches))
	for i, match := range matches {
		result[i] = diff.LineMatch{StartCol: match.StartCol, EndCol: match.EndCol}
	}
	return result
}

func (m *DiffModal) highlightFn(content string, matches []diff.LineMatch, baseStyle lipgloss.Style) string {
	var searchMatches []search.Match
	for _, match := range matches {
		searchMatches = append(searchMatches, search.Match{
			Start: match.StartCol,
			End:   match.EndCol,
		})
	}
	h := search.NewHighlighter(m.styles.SearchMatch, baseStyle)
	return h.Highlight(content, searchMatches)
}

func (m *DiffModal) renderContent() {
	if m.parsedDiff == nil || len(m.parsedDiff.Files) == 0 {
		m.viewport.SetContent("No changes in this commit")
		return
	}

	if m.sideBySide {
		// Side-by-side rendering uses the Renderer from diff package
		contentStr := m.renderer.RenderSideBySide(m.rawDiff, m.viewport.Width())
		m.filePositions = make([]diff.FilePosition, len(m.parsedDiff.Files))
		m.viewport.SetContent(contentStr)
		m.diffSearch.SetContent(contentStr)
		return
	}

	contentStr, positions := diff.RenderContent(
		m.parsedDiff,
		m.expanded,
		m.focusedFile,
		m.contentStyles(),
		m.matchProvider,
		m.highlightFn,
	)
	m.filePositions = positions
	m.viewport.SetContent(contentStr)
	m.diffSearch.SetContent(contentStr)
}

// getVisibleFileIndex returns which file is currently most visible in the viewport
func (m *DiffModal) getVisibleFileIndex() int {
	if len(m.filePositions) == 0 {
		return 0
	}

	viewMiddle := m.viewport.YOffset() + m.viewport.Height()/2

	for i, pos := range m.filePositions {
		if viewMiddle >= pos.StartLine && viewMiddle < pos.EndLine {
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
	m.viewport.SetYOffset(m.filePositions[fileIdx].StartLine)
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

	hasFiles := m.parsedDiff != nil && len(m.parsedDiff.Files) > 0

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
		case "S":
			// Toggle side-by-side diff view
			m.sideBySide = !m.sideBySide
			m.renderContent()
			return m, nil
		case "enter", " ":
			// Toggle focused file expansion
			if hasFiles && m.focusedFile < len(m.parsedDiff.Files) {
				m.expanded[m.focusedFile] = !m.expanded[m.focusedFile]
				// Update allExpanded state
				m.allExpanded = !diff.IsAllCollapsed(m.expanded)
				m.renderContent()
				// Scroll to keep file header visible
				m.scrollToFile(m.focusedFile)
			}
		case "a":
			// Toggle all files
			if hasFiles {
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
			if hasFiles && m.focusedFile < len(m.filePositions) {
				currentPos := m.filePositions[m.focusedFile]

				if m.expanded[m.focusedFile] {
					// Check if there's more content below in current file
					viewBottom := m.viewport.YOffset() + m.viewport.Height()
					if currentPos.EndLine > viewBottom {
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
			if hasFiles && m.focusedFile < len(m.filePositions) {
				currentPos := m.filePositions[m.focusedFile]

				if m.expanded[m.focusedFile] {
					// Check if we're not at the top of the file content
					if m.viewport.YOffset() > currentPos.StartLine {
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
			if hasFiles && m.focusedFile < len(m.parsedDiff.Files)-1 {
				m.focusedFile++
				m.renderContent()
				m.scrollToFile(m.focusedFile)
			}
		case "K":
			// Always move to previous file (Shift+k)
			if hasFiles && m.focusedFile > 0 {
				m.focusedFile--
				m.renderContent()
				m.scrollToFile(m.focusedFile)
			}
		case "}":
			// Jump to next file
			if hasFiles && m.focusedFile < len(m.parsedDiff.Files)-1 {
				m.focusedFile++
				m.renderContent()
				m.scrollToFile(m.focusedFile)
			}
		case "{":
			// Jump to previous file
			if hasFiles && m.focusedFile > 0 {
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
			if hasFiles {
				m.focusedFile = 0
			}
			m.renderContent()
		case "G":
			m.viewport.GotoBottom()
			if hasFiles {
				m.focusedFile = len(m.parsedDiff.Files) - 1
				m.renderContent()
			}
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

	modeStr := "unified"
	if m.sideBySide {
		modeStr = "side-by-side"
	}
	header := headerStyle.Render(fmt.Sprintf("%d files (%d expanded) [%s]", fileCount, expandedCount, modeStr))

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

	footer := footerStyle.Render("j/k nav • J/K file • {/} jump • enter toggle • a all • S split • / search • q close  ") + scrollbar

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
