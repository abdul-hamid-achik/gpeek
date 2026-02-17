package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	MinWidth  = 80
	MinHeight = 24
)

type Layout struct {
	Width  int
	Height int

	LeftWidth    int
	RightWidth   int
	TopHeight    int
	BottomHeight int

	FilesWidth    int
	FilesHeight   int
	BranchWidth   int
	BranchHeight  int
	CommitsWidth  int
	CommitsHeight int
	PreviewWidth  int
	PreviewHeight int

	StatusHeight int
}

func NewLayout(width, height int) *Layout {
	l := &Layout{
		Width:        width,
		Height:       height,
		StatusHeight: 1,
	}
	l.Calculate()
	return l
}

func (l *Layout) Calculate() {
	if l.Width < MinWidth {
		l.Width = MinWidth
	}
	if l.Height < MinHeight {
		l.Height = MinHeight
	}

	l.LeftWidth = l.Width / 4
	l.RightWidth = l.Width - l.LeftWidth

	availableHeight := l.Height - l.StatusHeight
	l.TopHeight = availableHeight / 2
	l.BottomHeight = availableHeight - l.TopHeight

	l.FilesWidth = l.LeftWidth
	l.FilesHeight = l.TopHeight

	l.BranchWidth = l.LeftWidth
	l.BranchHeight = l.BottomHeight

	l.CommitsWidth = l.RightWidth
	l.CommitsHeight = l.TopHeight

	l.PreviewWidth = l.RightWidth
	l.PreviewHeight = l.BottomHeight
}

func (l *Layout) SetSize(width, height int) {
	l.Width = width
	l.Height = height
	l.Calculate()
}

func (l *Layout) IsTooSmall() bool {
	return l.Width < MinWidth || l.Height < MinHeight
}

func (l *Layout) TooSmallMessage() string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#BF616A")).
		Render("Terminal too small. Minimum size: 80x24")
}

type PanelDimensions struct {
	Width       int
	Height      int
	InnerWidth  int
	InnerHeight int
}

func (l *Layout) FilesDimensions() PanelDimensions {
	return PanelDimensions{
		Width:       l.FilesWidth,
		Height:      l.FilesHeight,
		InnerWidth:  l.FilesWidth - 4,
		InnerHeight: l.FilesHeight - 2,
	}
}

func (l *Layout) BranchesDimensions() PanelDimensions {
	return PanelDimensions{
		Width:       l.BranchWidth,
		Height:      l.BranchHeight,
		InnerWidth:  l.BranchWidth - 4,
		InnerHeight: l.BranchHeight - 2,
	}
}

func (l *Layout) CommitsDimensions() PanelDimensions {
	return PanelDimensions{
		Width:       l.CommitsWidth,
		Height:      l.CommitsHeight,
		InnerWidth:  l.CommitsWidth - 4,
		InnerHeight: l.CommitsHeight - 2,
	}
}

func (l *Layout) PreviewDimensions() PanelDimensions {
	return PanelDimensions{
		Width:       l.PreviewWidth,
		Height:      l.PreviewHeight,
		InnerWidth:  l.PreviewWidth - 4,
		InnerHeight: l.PreviewHeight - 2,
	}
}

type FocusedPanel int

const (
	PanelFiles FocusedPanel = iota
	PanelBranches
	PanelCommits
	PanelPreview
)

func (f FocusedPanel) Next() FocusedPanel {
	return (f + 1) % 4
}

func (f FocusedPanel) Prev() FocusedPanel {
	if f == 0 {
		return PanelPreview
	}
	return f - 1
}

func (f FocusedPanel) String() string {
	switch f {
	case PanelFiles:
		return "Files"
	case PanelBranches:
		return "Branches"
	case PanelCommits:
		return "Commits"
	case PanelPreview:
		return "Preview"
	default:
		return ""
	}
}

func RenderBorder(content, title string, width, height int, focused bool, styles *Styles) string {
	if width < 4 {
		width = 4
	}
	if height < 3 {
		height = 3
	}

	style := styles.Panel
	titleStyle := styles.PanelTitle

	if focused {
		style = styles.PanelFocused
		titleStyle = styles.PanelTitleFocus
	}

	style = style.Width(width - 2).Height(height - 2)

	titleRendered := titleStyle.Render(title)
	titleWidth := lipgloss.Width(titleRendered)

	topBorder := "─" + titleRendered + strings.Repeat("─", max(0, width-titleWidth-4))

	bordered := style.Render(content)
	lines := strings.Split(bordered, "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		runeFirst := []rune(firstLine)
		if len(runeFirst) > 2 {
			newFirst := string(runeFirst[0]) + topBorder
			remaining := width - lipgloss.Width(newFirst) - 1
			if remaining > 0 {
				newFirst += strings.Repeat("─", remaining)
			}
			newFirst += string(runeFirst[len(runeFirst)-1])
			lines[0] = newFirst
		}
		bordered = strings.Join(lines, "\n")
	}

	return bordered
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
