package modals

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/abdul-hamid-achik/gpeek/internal/ui/panels"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CommitModal struct {
	BaseModal
	styles   *ui.Styles
	textarea textarea.Model
	staged   []panels.FileEntry
	onCommit func(string, bool) tea.Cmd
	err      string
	isAmend  bool
	lastCommitMsg string
}

func NewCommitModal(styles *ui.Styles, staged []panels.FileEntry, lastCommitMsg string, onCommit func(string, bool) tea.Cmd) *CommitModal {
	ta := textarea.New()
	ta.Placeholder = "Enter commit message..."
	ta.Focus()
	ta.CharLimit = 500
	ta.SetWidth(60)
	ta.SetHeight(5)

	return &CommitModal{
		styles:        styles,
		textarea:      ta,
		staged:        staged,
		onCommit:      onCommit,
		lastCommitMsg: lastCommitMsg,
	}
}

func (m *CommitModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return nil, nil
		case "ctrl+a":
			// Toggle amend mode
			m.isAmend = !m.isAmend
			if m.isAmend && m.lastCommitMsg != "" && strings.TrimSpace(m.textarea.Value()) == "" {
				// Pre-fill with last commit message
				m.textarea.SetValue(strings.TrimSpace(m.lastCommitMsg))
			}
			return m, nil
		case "ctrl+s", "ctrl+enter":
			message := strings.TrimSpace(m.textarea.Value())
			if message == "" {
				m.err = "Commit message cannot be empty"
				return m, nil
			}
			if m.onCommit != nil {
				return nil, m.onCommit(message, m.isAmend)
			}
			return nil, nil
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m *CommitModal) View() string {
	titleText := " Commit "
	if m.isAmend {
		titleText = " Amend Commit "
	}
	title := m.styles.ModalTitle.Render(titleText)

	var stagedList []string
	for _, f := range m.staged {
		stagedList = append(stagedList, "  "+f.Path)
	}

	stagedSection := m.styles.Bold.Render(fmt.Sprintf("Staged files (%d):", len(m.staged))) + "\n"
	if len(stagedList) > 5 {
		stagedSection += strings.Join(stagedList[:5], "\n")
		stagedSection += fmt.Sprintf("\n  ... and %d more", len(stagedList)-5)
	} else {
		stagedSection += strings.Join(stagedList, "\n")
	}

	messageLabel := m.styles.Bold.Render("Commit message:")

	var amendWarning string
	if m.isAmend {
		amendWarning = "\n" + m.styles.Warning.Render("⚠ AMEND: This will rewrite history")
	}

	var errLine string
	if m.err != "" {
		errLine = "\n" + m.styles.Error.Render(m.err)
	}

	footer := m.styles.Dim.Render("Ctrl+S to commit • Ctrl+A to toggle amend • Esc to cancel")

	body := lipgloss.JoinVertical(lipgloss.Left,
		stagedSection,
		amendWarning,
		"",
		messageLabel,
		m.textarea.View(),
		errLine,
		"",
		footer,
	)

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
