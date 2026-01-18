package modals

import (
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HelpModal struct {
	BaseModal
	styles   *ui.Styles
	viewport viewport.Model
	bindings [][]key.Binding
}

func NewHelpModal(styles *ui.Styles, bindings [][]key.Binding) *HelpModal {
	vp := viewport.New(60, 20)

	m := &HelpModal{
		styles:   styles,
		viewport: vp,
		bindings: bindings,
	}

	m.viewport.SetContent(m.renderContent())
	return m
}

func (m *HelpModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "?":
			return nil, nil
		case "j", "down":
			m.viewport.ScrollDown(1)
		case "k", "up":
			m.viewport.ScrollUp(1)
		case "ctrl+d":
			m.viewport.HalfPageDown()
		case "ctrl+u":
			m.viewport.HalfPageUp()
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *HelpModal) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Primary)).
		Background(lipgloss.Color(m.styles.Theme.Background)).
		Bold(true).
		Padding(0, 1)

	title := titleStyle.Render("Help")

	content := m.viewport.View()

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Muted)).
		Background(lipgloss.Color(m.styles.Theme.Background))

	footer := footerStyle.Render("Press ? or Esc to close")

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		content,
		"",
		footer,
	)

	return m.styles.Modal.Render(body)
}

func (m *HelpModal) renderContent() string {
	var sections []string

	categories := []string{
		"Navigation",
		"Panel Focus",
		"Staging",
		"Remote",
		"Branch",
		"Actions",
		"General",
	}

	for i, bindings := range m.bindings {
		if i >= len(categories) {
			break
		}

		var lines []string
		lines = append(lines, m.styles.Bold.Render(categories[i]))

		for _, b := range bindings {
			keys := b.Help().Key
			desc := b.Help().Desc
			line := m.styles.HelpKey.Render(padRight(keys, 12)) + " " +
				m.styles.HelpDesc.Render(desc)
			lines = append(lines, "  "+line)
		}

		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
