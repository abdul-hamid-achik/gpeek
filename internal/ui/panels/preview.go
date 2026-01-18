package panels

import (
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type PreviewPanel struct {
	BasePanel
	styles   *ui.Styles
	viewport viewport.Model

	content     string
	rawDiff     string
	highlighted bool
}

func NewPreviewPanel(styles *ui.Styles) *PreviewPanel {
	vp := viewport.New(0, 0)
	return &PreviewPanel{
		styles:   styles,
		viewport: vp,
	}
}

func (p *PreviewPanel) SetContent(content string) {
	p.content = content
	p.rawDiff = ""
	p.highlighted = false
	p.viewport.SetContent(content)
	p.viewport.GotoTop()
}

func (p *PreviewPanel) SetDiff(diffContent string) {
	p.rawDiff = diffContent
	p.highlighted = true

	renderer := diff.NewRenderer(p.styles)
	rendered := renderer.Render(diffContent, p.width)
	p.content = rendered
	p.viewport.SetContent(rendered)
	p.viewport.GotoTop()
}

func (p *PreviewPanel) Update(msg tea.Msg) tea.Cmd {
	if !p.focused {
		return nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			p.viewport.ScrollDown(1)
		case "k", "up":
			p.viewport.ScrollUp(1)
		case "g":
			p.viewport.GotoTop()
		case "G":
			p.viewport.GotoBottom()
		case "ctrl+d":
			p.viewport.HalfPageDown()
		case "ctrl+u":
			p.viewport.HalfPageUp()
		case "ctrl+f":
			p.viewport.PageDown()
		case "ctrl+b":
			p.viewport.PageUp()
		default:
			p.viewport, cmd = p.viewport.Update(msg)
		}
	default:
		p.viewport, cmd = p.viewport.Update(msg)
	}

	return cmd
}

func (p *PreviewPanel) View() string {
	if p.content == "" {
		return p.styles.Dim.Render("Select a file or commit to preview")
	}

	content := p.viewport.View()

	lines := strings.Split(content, "\n")
	for len(lines) < p.height {
		lines = append(lines, "")
	}
	if len(lines) > p.height {
		lines = lines[:p.height]
	}

	return strings.Join(lines, "\n")
}

func (p *PreviewPanel) SetSize(width, height int) {
	p.BasePanel.SetSize(width, height)
	p.viewport.Width = width
	p.viewport.Height = height

	if p.rawDiff != "" && p.highlighted {
		renderer := diff.NewRenderer(p.styles)
		rendered := renderer.Render(p.rawDiff, width)
		p.content = rendered
		p.viewport.SetContent(rendered)
	}
}

func (p *PreviewPanel) ScrollPercent() float64 {
	return p.viewport.ScrollPercent()
}
