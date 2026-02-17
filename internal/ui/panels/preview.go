package panels

import (
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/search"
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
	parsedDiff  *diff.Diff
	expanded    map[int]bool
	allExpanded bool
	focusedFile int

	// File position tracking for navigation
	filePositions []diff.FilePosition

	// Diff search
	diffSearch *uisearch.DiffSearch
}

func NewPreviewPanel(styles *ui.Styles) *PreviewPanel {
	vp := viewport.New(0, 0)
	return &PreviewPanel{
		styles:        styles,
		viewport:      vp,
		filePositions: make([]diff.FilePosition, 0),
		diffSearch:    uisearch.NewDiffSearch(styles),
	}
}

func (p *PreviewPanel) SetContent(content string) {
	p.content = content
	p.rawDiff = ""
	p.highlighted = false
	p.parsedDiff = nil
	p.expanded = nil
	p.filePositions = make([]diff.FilePosition, 0)
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
	p.filePositions = make([]diff.FilePosition, len(p.parsedDiff.Files))
	for i := range p.parsedDiff.Files {
		p.expanded[i] = false
	}
	p.allExpanded = false
	p.focusedFile = 0

	p.renderContent()
	p.viewport.GotoTop()
	p.diffSearch.SetContent(p.content)
}

func (p *PreviewPanel) contentStyles() diff.ContentStyles {
	return diff.ContentStyles{
		DiffMeta:    p.styles.DiffMeta,
		DiffHunk:    p.styles.DiffHunk,
		DiffAdd:     p.styles.DiffAdd,
		DiffRemove:  p.styles.DiffRemove,
		DiffContext:  p.styles.DiffContext,
		SearchMatch: p.styles.SearchMatch,
		FocusedFile: lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.styles.Theme.Background)).
			Background(lipgloss.Color(p.styles.Theme.Primary)).
			Bold(true),
	}
}

func (p *PreviewPanel) matchProvider(lineNum int) []diff.LineMatch {
	matches := p.diffSearch.GetLineMatches(lineNum)
	if len(matches) == 0 {
		return nil
	}
	result := make([]diff.LineMatch, len(matches))
	for i, match := range matches {
		result[i] = diff.LineMatch{StartCol: match.StartCol, EndCol: match.EndCol}
	}
	return result
}

func (p *PreviewPanel) highlightFn(content string, matches []diff.LineMatch, baseStyle lipgloss.Style) string {
	var searchMatches []search.Match
	for _, match := range matches {
		searchMatches = append(searchMatches, search.Match{
			Start: match.StartCol,
			End:   match.EndCol,
		})
	}
	h := search.NewHighlighter(p.styles.SearchMatch, baseStyle)
	return h.Highlight(content, searchMatches)
}

func (p *PreviewPanel) renderContent() {
	if p.parsedDiff == nil || len(p.parsedDiff.Files) == 0 {
		p.content = "No changes"
		p.viewport.SetContent(p.content)
		return
	}

	contentStr, positions := diff.RenderContent(
		p.parsedDiff,
		p.expanded,
		p.focusedFile,
		p.contentStyles(),
		p.matchProvider,
		p.highlightFn,
	)
	p.filePositions = positions
	p.content = contentStr
	p.viewport.SetContent(p.content)
}

// getVisibleFileIndex returns which file is currently most visible in the viewport
func (p *PreviewPanel) getVisibleFileIndex() int {
	if len(p.filePositions) == 0 {
		return 0
	}

	viewMiddle := p.viewport.YOffset + p.viewport.Height/2

	for i, pos := range p.filePositions {
		if viewMiddle >= pos.StartLine && viewMiddle < pos.EndLine {
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
	p.viewport.SetYOffset(p.filePositions[fileIdx].StartLine)
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
				p.allExpanded = !diff.IsAllCollapsed(p.expanded)
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
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 && p.focusedFile < len(p.filePositions) {
				currentPos := p.filePositions[p.focusedFile]

				if p.expanded[p.focusedFile] {
					// Check if there's more content below in current file
					viewBottom := p.viewport.YOffset + p.viewport.Height
					if currentPos.EndLine > viewBottom {
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
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 && p.focusedFile < len(p.filePositions) {
				currentPos := p.filePositions[p.focusedFile]

				if p.expanded[p.focusedFile] {
					// Check if we're not at the top of the file content
					if p.viewport.YOffset > currentPos.StartLine {
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
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				p.focusedFile = p.getVisibleFileIndex()
				p.renderContent()
			}
		case "ctrl+u":
			p.viewport.HalfPageUp()
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				p.focusedFile = p.getVisibleFileIndex()
				p.renderContent()
			}
		case "g":
			p.viewport.GotoTop()
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				p.focusedFile = 0
				p.renderContent()
			}
		case "G":
			p.viewport.GotoBottom()
			if p.parsedDiff != nil && len(p.parsedDiff.Files) > 0 {
				p.focusedFile = len(p.parsedDiff.Files) - 1
				p.renderContent()
			}
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
		msg := p.styles.Dim.Render("Select a file or commit to preview")
		if p.width > 0 && p.height > 0 {
			return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center, msg)
		}
		return msg
	}

	// Calculate content height (reserve space for search bar if active)
	contentHeight := p.height
	if p.diffSearch.IsActive() || p.diffSearch.HasSearch() {
		contentHeight--
	}

	if contentHeight <= 0 {
		contentHeight = 1
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
