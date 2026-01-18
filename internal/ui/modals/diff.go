package modals

import (
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DiffViewMode int

const (
	DiffViewUnified DiffViewMode = iota
	DiffViewSplit
)

type DiffModal struct {
	BaseModal
	styles   *ui.Styles
	viewport viewport.Model
	filename string
	rawDiff  string
	mode     DiffViewMode
}

func NewDiffModal(styles *ui.Styles, filename, diffContent string, width, height int) *DiffModal {
	vp := viewport.New(width-4, height-6)

	m := &DiffModal{
		styles:   styles,
		viewport: vp,
		filename: filename,
		rawDiff:  diffContent,
		mode:     DiffViewUnified,
	}
	m.width = width
	m.height = height

	m.renderDiff()
	return m
}

func (m *DiffModal) renderDiff() {
	renderer := diff.NewRenderer(m.styles)
	var content string

	if m.mode == DiffViewSplit {
		content = renderer.RenderSideBySide(m.rawDiff, m.viewport.Width)
	} else {
		content = renderer.Render(m.rawDiff, m.viewport.Width)
	}

	m.viewport.SetContent(content)
}

func (m *DiffModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return nil, nil
		case "v":
			if m.mode == DiffViewUnified {
				m.mode = DiffViewSplit
			} else {
				m.mode = DiffViewUnified
			}
			m.renderDiff()
		case "j", "down":
			m.viewport.ScrollDown(1)
		case "k", "up":
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
	title := m.styles.ModalTitle.Render(" " + m.filename + " ")

	modeIndicator := "unified"
	if m.mode == DiffViewSplit {
		modeIndicator = "split"
	}

	header := m.styles.Dim.Render("View: " + modeIndicator + " (v to toggle)")

	content := m.viewport.View()

	scrollPercent := int(m.viewport.ScrollPercent() * 100)
	footer := m.styles.Dim.Render(
		"j/k scroll • g/G top/bottom • q close  " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Render(
				strings.Repeat("█", scrollPercent/10)+strings.Repeat("░", 10-scrollPercent/10),
			),
	)

	body := lipgloss.JoinVertical(lipgloss.Left,
		header,
		content,
		footer,
	)

	modal := m.styles.Modal.
		Width(m.width).
		Height(m.height).
		Render(body)

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
