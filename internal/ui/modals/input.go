package modals

import (
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	modalWidth  int
}

func NewInputModal(styles *ui.Styles, title, prompt, placeholder string, onSubmit func(string) tea.Cmd) *InputModal {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 100
	ti.SetWidth(40)

	return &InputModal{
		styles:      styles,
		title:       title,
		prompt:      prompt,
		textinput:   ti,
		onSubmit:    onSubmit,
		placeholder: placeholder,
		modalWidth:  40,
	}
}

// SetTerminalWidth adjusts the modal width based on terminal size.
func (m *InputModal) SetTerminalWidth(termWidth int) {
	w := termWidth - 10
	if w > 60 {
		w = 60
	}
	if w < 30 {
		w = 30
	}
	m.modalWidth = w
	// Account for modal padding (2 on each side) and border (1 on each side)
	m.textinput.SetWidth(w - 6)
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

	modal := m.styles.Modal.Width(m.modalWidth).Render(body)

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
