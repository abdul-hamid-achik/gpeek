package modals

import (
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfirmModal struct {
	BaseModal
	styles    *ui.Styles
	title     string
	message   string
	onConfirm func() tea.Cmd
	focused   int
}

func NewConfirmModal(styles *ui.Styles, title, message string, onConfirm func() tea.Cmd) *ConfirmModal {
	return &ConfirmModal{
		styles:    styles,
		title:     title,
		message:   message,
		onConfirm: onConfirm,
		focused:   1,
	}
}

func (m *ConfirmModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "n", "N":
			return nil, nil
		case "y", "Y":
			// y/Y always confirms regardless of button focus
			if m.onConfirm != nil {
				return nil, m.onConfirm()
			}
			return nil, nil
		case "enter":
			if m.focused == 0 && m.onConfirm != nil {
				return nil, m.onConfirm()
			}
			if m.focused == 1 {
				return nil, nil
			}
		case "tab", "left", "right", "h", "l":
			m.focused = 1 - m.focused
		}
	}

	return m, nil
}

func (m *ConfirmModal) View() string {
	title := m.styles.ModalTitle.Render(" " + m.title + " ")

	message := m.styles.Base.Render(m.message)

	yesStyle := m.styles.Dim
	noStyle := m.styles.Dim

	if m.focused == 0 {
		yesStyle = m.styles.ListItemSelected
	} else {
		noStyle = m.styles.ListItemSelected
	}

	yes := yesStyle.Padding(0, 2).Render("Yes")
	no := noStyle.Padding(0, 2).Render("No")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yes, "  ", no)

	body := lipgloss.JoinVertical(lipgloss.Center,
		message,
		"",
		buttons,
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
