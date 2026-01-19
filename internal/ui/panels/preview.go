package panels

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	uisearch "github.com/abdul-hamid-achik/gpeek/internal/ui/search"
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

	// Diff search
	diffSearch *uisearch.DiffSearch
}

func NewPreviewPanel(styles *ui.Styles) *PreviewPanel {
	vp := viewport.New(0, 0)
	return &PreviewPanel{
		styles:     styles,
		viewport:   vp,
		diffSearch: uisearch.NewDiffSearch(styles),
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
	p.diffSearch.SetContent(content)
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
	p.diffSearch.SetContent(p.content)
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

// scrollToFocusedFile scrolls the viewport to show the focused file header
func (p *PreviewPanel) scrollToFocusedFile() {
	if p.parsedDiff == nil || len(p.parsedDiff.Files) == 0 {
		return
	}

	// Calculate the line number where the focused file header appears
	lineNum := 0
	for i := 0; i < p.focusedFile; i++ {
		lineNum++ // File header line
		if p.expanded[i] && !p.parsedDiff.Files[i].IsBinary {
			for _, hunk := range p.parsedDiff.Files[i].Hunks {
				lineNum++ // Hunk header
				lineNum += len(hunk.Lines)
			}
			lineNum++ // Extra blank line after expanded file
		}
	}

	// Scroll to the calculated line
	p.viewport.SetYOffset(lineNum)
}

func (p *PreviewPanel) SetSize(width, height int) {
	p.BasePanel.SetSize(width, height)
	p.viewport.Width = width
	p.viewport.Height = height
	p.diffSearch.SetWidth(width)

	// Re-render content when size changes (for diff content)
	if p.rawDiff != "" && p.highlighted && p.parsedDiff != nil {
		p.renderContent()
	}
}

func (p *PreviewPanel) Update(msg tea.Msg) tea.Cmd {
	if !p.focused {
		return nil
	}

	var cmd tea.Cmd

	// Handle diff search if active
	if p.diffSearch.IsActive() {
		searchCmd, scrollTo := p.diffSearch.Update(msg)
		if scrollTo >= 0 {
			p.viewport.SetYOffset(scrollTo)
		}
		return searchCmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			p.diffSearch.Activate()
			return nil
		case "n":
			if p.diffSearch.HasSearch() {
				scrollTo := p.diffSearch.NextMatch()
				if scrollTo >= 0 {
					p.viewport.SetYOffset(scrollTo)
				}
				return nil
			}
		case "N":
			if p.diffSearch.HasSearch() {
				scrollTo := p.diffSearch.PrevMatch()
				if scrollTo >= 0 {
					p.viewport.SetYOffset(scrollTo)
				}
				return nil
			}
		case "esc":
			if p.diffSearch.HasSearch() {
				p.diffSearch.Deactivate()
				return nil
			}
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
			// Always navigate between file headers
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				if p.focusedFile < len(p.parsedDiff.Files)-1 {
					p.focusedFile++
					p.renderContent()
					p.scrollToFocusedFile()
				}
			}
		case "k", "up":
			// Always navigate between file headers
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				if p.focusedFile > 0 {
					p.focusedFile--
					p.renderContent()
					p.scrollToFocusedFile()
				}
			}
		case "ctrl+j", "J":
			// Scroll viewport content
			p.viewport.ScrollDown(1)
		case "ctrl+k", "K":
			// Scroll viewport content
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

	// Calculate content height (reserve space for search bar if active)
	contentHeight := p.height
	if p.diffSearch.IsActive() || p.diffSearch.HasSearch() {
		contentHeight--
	}

	content := p.viewport.View()

	lines := strings.Split(content, "\n")
	for len(lines) < contentHeight {
		lines = append(lines, "")
	}
	if len(lines) > contentHeight {
		lines = lines[:contentHeight]
	}

	result := strings.Join(lines, "\n")

	// Add search bar if active
	if p.diffSearch.IsActive() || p.diffSearch.HasSearch() {
		result += "\n" + p.diffSearch.View()
	}

	return result
}

func (p *PreviewPanel) ScrollPercent() float64 {
	return p.viewport.ScrollPercent()
}

// IsSearching returns true if diff search is active
func (p *PreviewPanel) IsSearching() bool {
	return p.diffSearch.IsActive()
}
