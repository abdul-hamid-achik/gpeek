package modals

import (
	"fmt"
	"strings"
	"time"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StashMode int

const (
	StashModeList StashMode = iota
	StashModeCreate
	StashModePreview
)

type StashModal struct {
	BaseModal
	styles   *ui.Styles
	mode     StashMode
	stashes  []git.Stash
	cursor   int
	repo     *git.Repository
	width    int
	height   int

	// Create mode
	messageInput textinput.Model

	// Preview mode
	previewContent string
	previewScroll  int

	err string
}

func NewStashModal(styles *ui.Styles, stashes []git.Stash, repo *git.Repository, width, height int) *StashModal {
	msgInput := textinput.New()
	msgInput.Placeholder = "Stash message (optional)"
	msgInput.Width = 50

	return &StashModal{
		styles:       styles,
		stashes:      stashes,
		repo:         repo,
		messageInput: msgInput,
		width:        width,
		height:       height,
	}
}

func (m *StashModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case StashModeList:
			return m.updateListMode(msg)
		case StashModeCreate:
			return m.updateCreateMode(msg)
		case StashModePreview:
			return m.updatePreviewMode(msg)
		}
	}
	return m, nil
}

func (m *StashModal) updateListMode(msg tea.KeyMsg) (Modal, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		return nil, nil

	case "n":
		m.mode = StashModeCreate
		m.messageInput.Focus()
		m.err = ""
		return m, nil

	case "p":
		if len(m.stashes) > 0 {
			stash := m.stashes[m.cursor]
			if err := m.repo.StashPop(stash.Index); err != nil {
				m.err = err.Error()
			} else {
				m.refreshStashes()
			}
		}
		return m, nil

	case "a":
		if len(m.stashes) > 0 {
			stash := m.stashes[m.cursor]
			if err := m.repo.StashApply(stash.Index); err != nil {
				m.err = err.Error()
			} else {
				m.err = "Applied stash (kept in list)"
			}
		}
		return m, nil

	case "d":
		if len(m.stashes) > 0 {
			stash := m.stashes[m.cursor]
			if err := m.repo.StashDrop(stash.Index); err != nil {
				m.err = err.Error()
			} else {
				m.refreshStashes()
			}
		}
		return m, nil

	case "enter":
		if len(m.stashes) > 0 {
			stash := m.stashes[m.cursor]
			diff, err := m.repo.StashShow(stash.Index)
			if err != nil {
				m.err = err.Error()
			} else {
				m.previewContent = diff
				m.previewScroll = 0
				m.mode = StashModePreview
			}
		}
		return m, nil

	case "j", "down":
		if m.cursor < len(m.stashes)-1 {
			m.cursor++
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	}

	return m, nil
}

func (m *StashModal) updateCreateMode(msg tea.KeyMsg) (Modal, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = StashModeList
		m.messageInput.Reset()
		m.err = ""
		return m, nil

	case "enter":
		message := strings.TrimSpace(m.messageInput.Value())
		if err := m.repo.StashSave(message); err != nil {
			m.err = err.Error()
			return m, nil
		}

		m.refreshStashes()
		m.mode = StashModeList
		m.messageInput.Reset()
		m.err = ""
		return m, nil
	}

	var cmd tea.Cmd
	m.messageInput, cmd = m.messageInput.Update(msg)
	return m, cmd
}

func (m *StashModal) updatePreviewMode(msg tea.KeyMsg) (Modal, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.mode = StashModeList
		m.previewContent = ""
		return m, nil

	case "j", "down":
		m.previewScroll++

	case "k", "up":
		if m.previewScroll > 0 {
			m.previewScroll--
		}

	case "ctrl+d":
		m.previewScroll += 10

	case "ctrl+u":
		m.previewScroll -= 10
		if m.previewScroll < 0 {
			m.previewScroll = 0
		}
	}

	return m, nil
}

func (m *StashModal) refreshStashes() {
	stashes, _ := m.repo.StashList()
	m.stashes = stashes
	if m.cursor >= len(m.stashes) && m.cursor > 0 {
		m.cursor = len(m.stashes) - 1
	}
}

func (m *StashModal) View() string {
	var title string
	var body string

	switch m.mode {
	case StashModeCreate:
		title = m.styles.ModalTitle.Render(" New Stash ")
		body = m.renderCreateView()
	case StashModePreview:
		title = m.styles.ModalTitle.Render(" Stash Preview ")
		body = m.renderPreviewView()
	default:
		title = m.styles.ModalTitle.Render(" Stashes ")
		body = m.renderListView()
	}

	modal := m.styles.Modal.Render(body)

	lines := strings.Split(modal, "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
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
			lines[0] = string(runes)
		}
		modal = strings.Join(lines, "\n")
	}

	return modal
}

func (m *StashModal) renderListView() string {
	var lines []string

	if len(m.stashes) == 0 {
		lines = append(lines, m.styles.Dim.Render("No stashes"))
		lines = append(lines, "")
		lines = append(lines, m.styles.Dim.Render("Press n to create a new stash"))
	} else {
		for i, s := range m.stashes {
			line := m.renderStashItem(s, i == m.cursor)
			lines = append(lines, line)
		}
	}

	var errLine string
	if m.err != "" {
		errLine = "\n" + m.styles.Error.Render(m.err)
	}

	footer := m.styles.Dim.Render("n new • p pop • a apply • d drop • enter preview • q close")

	return lipgloss.JoinVertical(lipgloss.Left,
		strings.Join(lines, "\n"),
		errLine,
		"",
		footer,
	)
}

func (m *StashModal) renderStashItem(s git.Stash, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}

	// Truncate message if too long
	message := s.Message
	maxMsgLen := 50
	if len(message) > maxMsgLen {
		message = message[:maxMsgLen-3] + "..."
	}

	// Format time ago
	timeAgo := formatTimeAgo(s.Time)

	line := fmt.Sprintf("%sstash@{%d}: %s", prefix, s.Index, message)
	if s.Branch != "" {
		line = fmt.Sprintf("%sstash@{%d} [%s]: %s", prefix, s.Index, s.Branch, message)
	}

	// Add time info
	line += fmt.Sprintf(" (%s)", timeAgo)

	if selected {
		return m.styles.ListItemSelected.Render(line)
	}
	return m.styles.ListItem.Render(line)
}

func (m *StashModal) renderCreateView() string {
	msgLabel := m.styles.Bold.Render("Message (optional):")

	var errLine string
	if m.err != "" {
		errLine = "\n" + m.styles.Error.Render(m.err)
	}

	footer := m.styles.Dim.Render("Enter to create • Esc to cancel")

	return lipgloss.JoinVertical(lipgloss.Left,
		msgLabel,
		m.messageInput.View(),
		errLine,
		"",
		footer,
	)
}

func (m *StashModal) renderPreviewView() string {
	lines := strings.Split(m.previewContent, "\n")

	// Calculate visible area
	maxLines := m.height - 6
	if maxLines < 5 {
		maxLines = 5
	}

	start := m.previewScroll
	if start > len(lines)-maxLines {
		start = len(lines) - maxLines
	}
	if start < 0 {
		start = 0
	}

	end := start + maxLines
	if end > len(lines) {
		end = len(lines)
	}

	visibleLines := lines[start:end]
	content := strings.Join(visibleLines, "\n")

	scrollInfo := fmt.Sprintf("Line %d-%d of %d", start+1, end, len(lines))
	footer := m.styles.Dim.Render("j/k scroll • Ctrl+D/U page • q back")

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		"",
		m.styles.Dim.Render(scrollInfo),
		footer,
	)
}

func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else {
		return t.Format("Jan 2, 2006")
	}
}
