package panels

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PreviewPanel struct {
	BasePanel
	styles   *ui.Styles
	viewport viewport.Model

	content     string
	rawDiff     string
	highlighted bool

	// Collapsible file drawer state
	parsedDiff  *diff.Diff   // Parsed once at creation
	expanded    map[int]bool // File index → expanded state
	allExpanded bool         // Toggle all state
	focusedFile int          // Currently focused file index
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
	p.parsedDiff = nil
	p.expanded = nil
	p.viewport.SetContent(content)
	p.viewport.GotoTop()
}

func (p *PreviewPanel) SetDiff(diffContent string) {
	p.rawDiff = diffContent
	p.highlighted = true

	// Parse diff once
	p.parsedDiff = diff.Parse(diffContent)

	// Initialize all files as expanded
	p.expanded = make(map[int]bool)
	for i := range p.parsedDiff.Files {
		p.expanded[i] = true
	}
	p.allExpanded = true
	p.focusedFile = 0

	p.renderContent()
	p.viewport.GotoTop()
}

func (p *PreviewPanel) renderContent() {
	if p.parsedDiff == nil || len(p.parsedDiff.Files) == 0 {
		p.content = "No changes"
		p.viewport.SetContent(p.content)
		return
	}

	var content strings.Builder

	for i, file := range p.parsedDiff.Files {
		// Render file header with expand/collapse indicator
		indicator := "▶"
		if p.expanded[i] {
			indicator = "▼"
		}

		// Calculate stats for this file
		adds, dels := p.countFileChanges(file)

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

		if i == p.focusedFile {
			// Focused file header style
			headerStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(p.styles.Theme.Background)).
				Background(lipgloss.Color(p.styles.Theme.Primary)).
				Bold(true)
			content.WriteString(headerStyle.Render(header))
		} else {
			// Normal file header style
			content.WriteString(p.styles.DiffMeta.Render(header))
		}
		content.WriteString("\n")

		// Render file content if expanded (and not binary)
		if p.expanded[i] && !file.IsBinary {
			for _, hunk := range file.Hunks {
				content.WriteString(p.styles.DiffHunk.Render(hunk.Header))
				content.WriteString("\n")
				for _, line := range hunk.Lines {
					content.WriteString(p.renderLine(line))
					content.WriteString("\n")
				}
			}
			content.WriteString("\n")
		}
	}

	p.content = content.String()
	p.viewport.SetContent(p.content)
}

func (p *PreviewPanel) countFileChanges(file diff.FileDiff) (adds, dels int) {
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

func (p *PreviewPanel) renderLine(line diff.Line) string {
	switch line.Type {
	case diff.DiffAdd:
		return p.styles.DiffAdd.Render("+" + line.Content)
	case diff.DiffRemove:
		return p.styles.DiffRemove.Render("-" + line.Content)
	default:
		return p.styles.DiffContext.Render(" " + line.Content)
	}
}

func (p *PreviewPanel) isAllCollapsed() bool {
	for _, exp := range p.expanded {
		if exp {
			return false
		}
	}
	return true
}

func (p *PreviewPanel) Update(msg tea.Msg) tea.Cmd {
	if !p.focused {
		return nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			// Toggle focused file expansion
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				p.expanded[p.focusedFile] = !p.expanded[p.focusedFile]
				// Update allExpanded state
				p.allExpanded = !p.isAllCollapsed()
				p.renderContent()
			}
		case "a":
			// Toggle all files
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				p.allExpanded = !p.allExpanded
				for i := range p.parsedDiff.Files {
					p.expanded[i] = p.allExpanded
				}
				p.renderContent()
			}
		case "j", "down":
			if p.isAllCollapsed() && p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				// Navigate to next file when all collapsed
				if p.focusedFile < len(p.parsedDiff.Files)-1 {
					p.focusedFile++
					p.renderContent()
				}
			} else {
				// Scroll viewport when files are expanded
				p.viewport.ScrollDown(1)
			}
		case "k", "up":
			if p.isAllCollapsed() && p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				// Navigate to previous file when all collapsed
				if p.focusedFile > 0 {
					p.focusedFile--
					p.renderContent()
				}
			} else {
				// Scroll viewport when files are expanded
				p.viewport.ScrollUp(1)
			}
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

	// Re-render content when size changes (for diff content)
	if p.rawDiff != "" && p.highlighted && p.parsedDiff != nil {
		p.renderContent()
	}
}

func (p *PreviewPanel) ScrollPercent() float64 {
	return p.viewport.ScrollPercent()
}
