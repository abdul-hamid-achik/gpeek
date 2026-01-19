package modals

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DiffModal struct {
	BaseModal
	styles      *ui.Styles
	viewport    viewport.Model
	commitInfo  string           // Commit hash + message for title
	rawDiff     string           // Keep for reference
	parsedDiff  *diff.Diff       // Parsed once at creation
	expanded    map[int]bool     // File index → expanded state
	allExpanded bool             // Toggle all state
	focusedFile int              // Currently focused file index
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
		styles:      styles,
		viewport:    vp,
		commitInfo:  title,
		rawDiff:     diffContent,
		parsedDiff:  parsedDiff,
		expanded:    expanded,
		allExpanded: true,
		focusedFile: 0,
	}
	m.width = width
	m.height = height

	m.renderContent()
	return m
}

func (m *DiffModal) renderContent() {
	if len(m.parsedDiff.Files) == 0 {
		m.viewport.SetContent("No changes in this commit")
		return
	}

	var content strings.Builder

	for i, file := range m.parsedDiff.Files {
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

		// Render file content if expanded (and not binary)
		if m.expanded[i] && !file.IsBinary {
			for _, hunk := range file.Hunks {
				content.WriteString(m.styles.DiffHunk.Render(hunk.Header))
				content.WriteString("\n")
				for _, line := range hunk.Lines {
					content.WriteString(m.renderLine(line))
					content.WriteString("\n")
				}
			}
			content.WriteString("\n")
		}
	}

	m.viewport.SetContent(content.String())
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

func (m *DiffModal) renderLine(line diff.Line) string {
	switch line.Type {
	case diff.DiffAdd:
		return m.styles.DiffAdd.Render("+" + line.Content)
	case diff.DiffRemove:
		return m.styles.DiffRemove.Render("-" + line.Content)
	default:
		return m.styles.DiffContext.Render(" " + line.Content)
	}
}

func (m *DiffModal) isAllCollapsed() bool {
	for _, exp := range m.expanded {
		if exp {
			return false
		}
	}
	return true
}

func (m *DiffModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return nil, nil
		case "enter", " ":
			// Toggle focused file expansion
			if len(m.parsedDiff.Files) > 0 {
				m.expanded[m.focusedFile] = !m.expanded[m.focusedFile]
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
			if m.isAllCollapsed() && len(m.parsedDiff.Files) > 0 {
				// Navigate to next file when all collapsed
				if m.focusedFile < len(m.parsedDiff.Files)-1 {
					m.focusedFile++
					m.renderContent()
				}
			} else {
				// Scroll viewport when files are expanded
				m.viewport.ScrollDown(1)
			}
		case "k", "up":
			if m.isAllCollapsed() && len(m.parsedDiff.Files) > 0 {
				// Navigate to previous file when all collapsed
				if m.focusedFile > 0 {
					m.focusedFile--
					m.renderContent()
				}
			} else {
				// Scroll viewport when files are expanded
				m.viewport.ScrollUp(1)
			}
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

	footer := footerStyle.Render("j/k nav • enter toggle • a toggle all • q close  ") + scrollbar

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		header,
		"",
		content,
		"",
		footer,
	)

	return m.styles.Modal.
		Width(m.width).
		Height(m.height).
		Render(body)
}
