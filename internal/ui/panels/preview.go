package panels

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/search"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	uisearch "github.com/abdul-hamid-achik/gpeek/internal/ui/search"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type previewFilePosition struct {
	startLine int
	endLine   int
	expanded  bool
}

type PreviewPanel struct {
	BasePanel
	styles   *ui.Styles
	viewport viewport.Model

	content     string
	rawDiff     string
	highlighted bool

	// Collapsible file drawer state
	parsedDiff  *diff.Diff
	expanded    map[int]bool
	allExpanded bool
	focusedFile int

	// File position tracking for navigation
	filePositions []previewFilePosition

	// Diff search
	diffSearch *uisearch.DiffSearch
}

func NewPreviewPanel(styles *ui.Styles) *PreviewPanel {
	vp := viewport.New(0, 0)
	return &PreviewPanel{
		styles:        styles,
		viewport:      vp,
		filePositions: make([]previewFilePosition, 0),
		diffSearch:    uisearch.NewDiffSearch(styles),
	}
}

func (p *PreviewPanel) SetContent(content string) {
	p.content = content
	p.rawDiff = ""
	p.highlighted = false
	p.parsedDiff = nil
	p.expanded = nil
	p.filePositions = make([]previewFilePosition, 0)
	p.viewport.SetContent(content)
	p.viewport.GotoTop()
	p.diffSearch.SetContent(content)
}

func (p *PreviewPanel) SetDiff(diffContent string) {
	p.rawDiff = diffContent
	p.highlighted = true

	// Parse diff once
	p.parsedDiff = diff.Parse(diffContent)

	// Initialize all files as collapsed
	p.expanded = make(map[int]bool)
	p.filePositions = make([]previewFilePosition, len(p.parsedDiff.Files))
	for i := range p.parsedDiff.Files {
		p.expanded[i] = false
	}
	p.allExpanded = false
	p.focusedFile = 0

	p.renderContent()
	p.viewport.GotoTop()
	p.diffSearch.SetContent(p.content)
}

// fileLine represents a line in the file content
type previewFileLine struct {
	isHunk bool
	text   string
	line   diff.Line
}

func (p *PreviewPanel) renderContent() {
	if p.parsedDiff == nil || len(p.parsedDiff.Files) == 0 {
		p.content = "No changes"
		p.viewport.SetContent(p.content)
		return
	}

	var content strings.Builder
	lineNum := 0

	for i, file := range p.parsedDiff.Files {
		// Track file start position
		p.filePositions[i].startLine = lineNum
		p.filePositions[i].expanded = p.expanded[i]

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
		lineNum++

		// Render file content if expanded (and not binary)
		if p.expanded[i] && !file.IsBinary {
			// Collect all lines for this file
			var fileLines []previewFileLine
			for _, hunk := range file.Hunks {
				fileLines = append(fileLines, previewFileLine{isHunk: true, text: hunk.Header})
				for _, line := range hunk.Lines {
					fileLines = append(fileLines, previewFileLine{line: line})
				}
			}

			// Render ALL lines (no constraint)
			for j := 0; j < len(fileLines); j++ {
				fl := fileLines[j]
				if fl.isHunk {
					content.WriteString(p.styles.DiffHunk.Render(fl.text))
				} else {
					content.WriteString(p.renderLine(fl.line, lineNum))
				}
				content.WriteString("\n")
				lineNum++
			}

			content.WriteString("\n")
			lineNum++
		}

		// Track file end position
		p.filePositions[i].endLine = lineNum
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

func (p *PreviewPanel) renderLine(line diff.Line, lineNum int) string {
	prefix := " "
	var baseStyle lipgloss.Style

	switch line.Type {
	case diff.DiffAdd:
		prefix = "+"
		baseStyle = p.styles.DiffAdd
	case diff.DiffRemove:
		prefix = "-"
		baseStyle = p.styles.DiffRemove
	default:
		baseStyle = p.styles.DiffContext
	}

	content := line.Content

	// Apply search highlighting if matches exist
	matches := p.diffSearch.GetLineMatches(lineNum)
	if len(matches) > 0 {
		// Convert to search.Match format and highlight
		var searchMatches []search.Match
		for _, match := range matches {
			searchMatches = append(searchMatches, search.Match{
				Start: match.StartCol,
				End:   match.EndCol,
			})
		}
		h := search.NewHighlighter(p.styles.SearchMatch, baseStyle)
		return baseStyle.Render(prefix) + h.Highlight(content, searchMatches)
	}

	return baseStyle.Render(prefix + content)
}

func (p *PreviewPanel) isAllCollapsed() bool {
	for _, exp := range p.expanded {
		if exp {
			return false
		}
	}
	return true
}

// getVisibleFileIndex returns which file is currently most visible in the viewport
func (p *PreviewPanel) getVisibleFileIndex() int {
	viewMiddle := p.viewport.YOffset + p.viewport.Height/2

	for i, pos := range p.filePositions {
		if viewMiddle >= pos.startLine && viewMiddle < pos.endLine {
			return i
		}
	}

	// Default to last file if viewport is past all files
	return len(p.filePositions) - 1
}

// scrollToFile scrolls the viewport to show the file header
func (p *PreviewPanel) scrollToFile(fileIdx int) {
	if fileIdx < 0 || fileIdx >= len(p.filePositions) {
		return
	}

	// Scroll to the file header line
	p.viewport.SetYOffset(p.filePositions[fileIdx].startLine)
}

// ScrollToFileByName scrolls to a file by its name (used for Files panel sync)
func (p *PreviewPanel) ScrollToFileByName(filename string) {
	if p.parsedDiff == nil {
		return
	}

	for i, file := range p.parsedDiff.Files {
		name := file.NewName
		if name == "" || name == "/dev/null" {
			name = file.OldName
		}
		if name == filename {
			p.focusedFile = i
			p.renderContent()
			p.scrollToFile(i)
			return
		}
	}
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
				p.scrollToFile(p.focusedFile)
			}
		case "a":
			// Toggle all files
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				p.allExpanded = !p.allExpanded
				for i := range p.parsedDiff.Files {
					p.expanded[i] = p.allExpanded
				}
				p.renderContent()
				p.scrollToFile(p.focusedFile)
			}
		case "j", "down":
			// Hybrid navigation
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				currentPos := p.filePositions[p.focusedFile]

				if p.expanded[p.focusedFile] {
					// Check if there's more content below in current file
					viewBottom := p.viewport.YOffset + p.viewport.Height
					if currentPos.endLine > viewBottom {
						// More content below, scroll down
						p.viewport.ScrollDown(1)
						// Update focused file based on what's now visible
						p.focusedFile = p.getVisibleFileIndex()
						p.renderContent()
						return nil
					}
				}

				// At bottom of file or file collapsed - move to next file
				if p.focusedFile < len(p.parsedDiff.Files)-1 {
					p.focusedFile++
					p.renderContent()
					p.scrollToFile(p.focusedFile)
				}
			}
		case "k", "up":
			// Hybrid navigation
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				currentPos := p.filePositions[p.focusedFile]

				if p.expanded[p.focusedFile] {
					// Check if we're not at the top of the file content
					if p.viewport.YOffset > currentPos.startLine {
						// Can scroll up within file
						p.viewport.ScrollUp(1)
						// Update focused file based on what's now visible
						p.focusedFile = p.getVisibleFileIndex()
						p.renderContent()
						return nil
					}
				}

				// At top of file or file collapsed - move to prev file
				if p.focusedFile > 0 {
					p.focusedFile--
					p.renderContent()
					p.scrollToFile(p.focusedFile)
				}
			}
		case "J":
			// Always move to next file (Shift+j)
			if p.parsedDiff != nil && p.focusedFile < len(p.parsedDiff.Files)-1 {
				p.focusedFile++
				p.renderContent()
				p.scrollToFile(p.focusedFile)
			}
		case "K":
			// Always move to previous file (Shift+k)
			if p.parsedDiff != nil && p.focusedFile > 0 {
				p.focusedFile--
				p.renderContent()
				p.scrollToFile(p.focusedFile)
			}
		case "}":
			// Jump to next file
			if p.parsedDiff != nil && p.focusedFile < len(p.parsedDiff.Files)-1 {
				p.focusedFile++
				p.renderContent()
				p.scrollToFile(p.focusedFile)
			}
		case "{":
			// Jump to previous file
			if p.parsedDiff != nil && p.focusedFile > 0 {
				p.focusedFile--
				p.renderContent()
				p.scrollToFile(p.focusedFile)
			}
		case "ctrl+d":
			p.viewport.HalfPageDown()
			p.focusedFile = p.getVisibleFileIndex()
			p.renderContent()
		case "ctrl+u":
			p.viewport.HalfPageUp()
			p.focusedFile = p.getVisibleFileIndex()
			p.renderContent()
		case "g":
			p.viewport.GotoTop()
			p.focusedFile = 0
			p.renderContent()
		case "G":
			p.viewport.GotoBottom()
			p.focusedFile = len(p.parsedDiff.Files) - 1
			p.renderContent()
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
