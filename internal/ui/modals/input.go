package modals

import (
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InputModal struct {
	BaseModal
	styles      *ui.Styles
	title       string
	prompt      string
	textinput   textinput.Model
	onSubmit    func(string) tea.Cmd
	err         string
	placeholder string
}

func NewInputModal(styles *ui.Styles, title, prompt, placeholder string, onSubmit func(string) tea.Cmd) *InputModal {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	return &InputModal{
		styles:      styles,
		title:       title,
		prompt:      prompt,
		textinput:   ti,
		onSubmit:    onSubmit,
		placeholder: placeholder,
	}
}

func (m *InputModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return nil, nil
		case "enter":
			value := strings.TrimSpace(m.textinput.Value())
			if value == "" {
				m.err = "Value cannot be empty"
				return m, nil
			}
			if m.onSubmit != nil {
				return nil, m.onSubmit(value)
			}
			return nil, nil
		}
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m *InputModal) View() string {
	title := m.styles.ModalTitle.Render(" " + m.title + " ")

	promptLine := m.styles.Bold.Render(m.prompt)

	var errLine string
	if m.err != "" {
		errLine = "\n" + m.styles.Error.Render(m.err)
	}

	footer := m.styles.Dim.Render("Enter to confirm  Esc to cancel")

	body := lipgloss.JoinVertical(lipgloss.Left,
		promptLine,
		"",
		m.textinput.View(),
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
