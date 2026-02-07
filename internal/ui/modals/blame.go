package modals

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BlameModal struct {
	BaseModal
	styles   *ui.Styles
	filepath string
	lines    []git.BlameLine
	cursor   int
	scroll   int
	width    int
	height   int
	repo     *git.Repository

	// For showing commit diff
	showingDiff bool
	diffContent string
	diffScroll  int
}

func NewBlameModal(styles *ui.Styles, filepath string, lines []git.BlameLine, repo *git.Repository, width, height int) *BlameModal {
	return &BlameModal{
		styles:   styles,
		filepath: filepath,
		lines:    lines,
		repo:     repo,
		width:    width,
		height:   height,
	}
}

func (m *BlameModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showingDiff {
			return m.updateDiffMode(msg)
		}
		return m.updateBlameMode(msg)
	}
	return m, nil
}

func (m *BlameModal) updateBlameMode(msg tea.KeyMsg) (Modal, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		return nil, nil

	case "j", "down":
		if m.cursor < len(m.lines)-1 {
			m.cursor++
			m.ensureCursorVisible()
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.ensureCursorVisible()
		}

	case "ctrl+d":
		m.cursor += m.visibleLines() / 2
		if m.cursor >= len(m.lines) {
			m.cursor = len(m.lines) - 1
		}
		m.ensureCursorVisible()

	case "ctrl+u":
		m.cursor -= m.visibleLines() / 2
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.ensureCursorVisible()

	case "g":
		m.cursor = 0
		m.scroll = 0

	case "G":
		if len(m.lines) > 0 {
			m.cursor = len(m.lines) - 1
			m.ensureCursorVisible()
		}

	case "enter":
		if m.cursor < len(m.lines) && m.lines[m.cursor].Hash != "" {
			// Show diff for this commit
			hash := m.lines[m.cursor].Hash
			diff, err := m.repo.CommitDiff(hash)
			if err == nil {
				m.diffContent = diff
				m.diffScroll = 0
				m.showingDiff = true
			}
		}
	}

	return m, nil
}

func (m *BlameModal) updateDiffMode(msg tea.KeyMsg) (Modal, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.showingDiff = false
		m.diffContent = ""
		return m, nil

	case "j", "down":
		m.diffScroll++

	case "k", "up":
		if m.diffScroll > 0 {
			m.diffScroll--
		}

	case "ctrl+d":
		m.diffScroll += 10

	case "ctrl+u":
		m.diffScroll -= 10
		if m.diffScroll < 0 {
			m.diffScroll = 0
		}

	case "g":
		m.diffScroll = 0

	case "G":
		lines := strings.Split(m.diffContent, "\n")
		m.diffScroll = len(lines) - m.visibleLines()
		if m.diffScroll < 0 {
			m.diffScroll = 0
		}
	}

	return m, nil
}

func (m *BlameModal) visibleLines() int {
	return m.height - 5 // Account for title, footer, etc.
}

func (m *BlameModal) ensureCursorVisible() {
	visible := m.visibleLines()
	if visible <= 0 {
		return
	}

	// Scroll down if cursor is below visible area
	if m.cursor >= m.scroll+visible {
		m.scroll = m.cursor - visible + 1
	}

	// Scroll up if cursor is above visible area
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
}

func (m *BlameModal) View() string {
	if m.showingDiff {
		return m.renderDiffView()
	}
	return m.renderBlameView()
}

func (m *BlameModal) renderBlameView() string {
	title := m.styles.ModalTitle.Render(fmt.Sprintf(" Blame: %s ", m.filepath))

	var lines []string
	visible := m.visibleLines()

	// Ensure scroll bounds
	if m.scroll < 0 {
		m.scroll = 0
	}
	maxScroll := len(m.lines) - visible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scroll > maxScroll {
		m.scroll = maxScroll
	}

	end := m.scroll + visible
	if end > len(m.lines) {
		end = len(m.lines)
	}

	// Calculate column widths
	hashWidth := 8
	authorWidth := 12
	dateWidth := 10
	contentWidth := m.width - hashWidth - authorWidth - dateWidth - 10

	if contentWidth < 20 {
		contentWidth = 20
	}

	// Header
	header := fmt.Sprintf("%-*s %-*s %-*s  %s",
		hashWidth, "Hash",
		authorWidth, "Author",
		dateWidth, "Date",
		"Content")
	lines = append(lines, m.styles.Bold.Render(header))
	lines = append(lines, strings.Repeat("─", m.width-4))

	for i := m.scroll; i < end; i++ {
		bl := m.lines[i]
		selected := i == m.cursor

		// Format hash
		hash := "        "
		if bl.Hash != "" && len(bl.Hash) >= 7 {
			hash = bl.Hash[:7] + " "
		}

		// Format author (truncate if needed)
		author := bl.Author
		if len(author) > authorWidth {
			author = author[:authorWidth-1] + "…"
		}
		author = fmt.Sprintf("%-*s", authorWidth, author)

		// Format date
		date := "          "
		if !bl.Time.IsZero() {
			date = bl.Time.Format("2006-01-02")
		}

		// Format content (truncate if needed)
		content := bl.Content
		if len(content) > contentWidth {
			content = content[:contentWidth-1] + "…"
		}

		// Line number prefix
		lineNum := fmt.Sprintf("%4d", bl.LineNum)

		line := fmt.Sprintf("%s %s %s %s  %s", lineNum, hash, author, date, content)

		if selected {
			line = m.styles.ListItemSelected.Render(line)
		} else {
			line = m.styles.ListItem.Render(line)
		}

		lines = append(lines, line)
	}

	// Scroll indicator
	scrollInfo := fmt.Sprintf("Line %d/%d", m.cursor+1, len(m.lines))

	footer := m.styles.Dim.Render("j/k navigate • enter view commit • g/G top/bottom • q close")

	body := lipgloss.JoinVertical(lipgloss.Left,
		strings.Join(lines, "\n"),
		"",
		m.styles.Dim.Render(scrollInfo),
		footer,
	)

	modal := m.styles.Modal.Width(m.width).Render(body)

	// Overlay title on border
	modalLines := strings.Split(modal, "\n")
	if len(modalLines) > 0 {
		firstLine := modalLines[0]
		titleWidth := lipgloss.Width(title)
		borderStart := 2

		if len(firstLine) > borderStart+titleWidth {
			runes := []rune(firstLine)
			titleRunes := []rune(title)
			for i, r := range titleRunes {
				if borderStart+i < len(runes) {
					runes[borderStart+i] = r
				}
			}
			modalLines[0] = string(runes)
		}
		modal = strings.Join(modalLines, "\n")
	}

	return modal
}

func (m *BlameModal) renderDiffView() string {
	// Get the hash for the title
	hash := ""
	if m.cursor < len(m.lines) && m.lines[m.cursor].Hash != "" {
		h := m.lines[m.cursor].Hash
		if len(h) > 7 {
			h = h[:7]
		}
		hash = h
	}
	title := m.styles.ModalTitle.Render(fmt.Sprintf(" Commit %s ", hash))

	lines := strings.Split(m.diffContent, "\n")
	visible := m.visibleLines()

	// Ensure scroll bounds
	if m.diffScroll < 0 {
		m.diffScroll = 0
	}
	maxScroll := len(lines) - visible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.diffScroll > maxScroll {
		m.diffScroll = maxScroll
	}

	end := m.diffScroll + visible
	if end > len(lines) {
		end = len(lines)
	}

	visibleContent := strings.Join(lines[m.diffScroll:end], "\n")

	scrollInfo := fmt.Sprintf("Line %d-%d of %d", m.diffScroll+1, end, len(lines))
	footer := m.styles.Dim.Render("j/k scroll • Ctrl+D/U page • q back to blame")

	body := lipgloss.JoinVertical(lipgloss.Left,
		visibleContent,
		"",
		m.styles.Dim.Render(scrollInfo),
		footer,
	)

	modal := m.styles.Modal.Width(m.width).Render(body)

	// Overlay title
	modalLines := strings.Split(modal, "\n")
	if len(modalLines) > 0 {
		firstLine := modalLines[0]
		titleWidth := lipgloss.Width(title)
		borderStart := 2

		if len(firstLine) > borderStart+titleWidth {
			runes := []rune(firstLine)
			titleRunes := []rune(title)
			for i, r := range titleRunes {
				if borderStart+i < len(runes) {
					runes[borderStart+i] = r
				}
			}
			modalLines[0] = string(runes)
		}
		modal = strings.Join(modalLines, "\n")
	}

	return modal
}
