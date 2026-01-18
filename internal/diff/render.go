package diff

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/ui"
)

type Renderer struct {
	styles *ui.Styles
}

func NewRenderer(styles *ui.Styles) *Renderer {
	return &Renderer{styles: styles}
}

func (r *Renderer) Render(diffContent string, width int) string {
	if diffContent == "" {
		return r.styles.Dim.Render("No diff content")
	}

	parsed := Parse(diffContent)
	var result strings.Builder

	for i, file := range parsed.Files {
		if i > 0 {
			result.WriteString("\n")
		}

		result.WriteString(r.renderFileHeader(file))
		result.WriteString("\n")

		for _, hunk := range file.Hunks {
			result.WriteString(r.renderHunkHeader(hunk))
			result.WriteString("\n")

			for _, line := range hunk.Lines {
				result.WriteString(r.renderLine(line, width))
				result.WriteString("\n")
			}
		}
	}

	return result.String()
}

func (r *Renderer) renderFileHeader(file FileDiff) string {
	var header string
	if file.IsNew {
		header = fmt.Sprintf("new file: %s", file.NewName)
	} else if file.IsDelete {
		header = fmt.Sprintf("deleted: %s", file.OldName)
	} else if file.IsRename {
		header = fmt.Sprintf("renamed: %s → %s", file.OldName, file.NewName)
	} else if file.IsBinary {
		header = fmt.Sprintf("binary: %s", file.NewName)
	} else {
		header = fmt.Sprintf("modified: %s", file.NewName)
	}

	return r.styles.DiffMeta.Bold(true).Render(header)
}

func (r *Renderer) renderHunkHeader(hunk Hunk) string {
	return r.styles.DiffHunk.Render(hunk.Header)
}

func (r *Renderer) renderLine(line Line, width int) string {
	lineNumWidth := 4
	prefix := " "
	style := r.styles.DiffContext

	switch line.Type {
	case DiffAdd:
		prefix = "+"
		style = r.styles.DiffAdd
	case DiffRemove:
		prefix = "-"
		style = r.styles.DiffRemove
	case DiffContext:
		prefix = " "
		style = r.styles.DiffContext
	}

	var lineNum string
	switch line.Type {
	case DiffAdd:
		lineNum = fmt.Sprintf("%*d", lineNumWidth, line.NewNumber)
	case DiffRemove:
		lineNum = fmt.Sprintf("%*d", lineNumWidth, line.OldNumber)
	default:
		if line.OldNumber > 0 {
			lineNum = fmt.Sprintf("%*d", lineNumWidth, line.OldNumber)
		} else {
			lineNum = strings.Repeat(" ", lineNumWidth)
		}
	}

	content := line.Content
	maxContentWidth := width - lineNumWidth - 3
	if len(content) > maxContentWidth && maxContentWidth > 0 {
		content = content[:maxContentWidth]
	}

	return fmt.Sprintf("%s %s%s",
		r.styles.LineNumber.Render(lineNum),
		style.Render(prefix),
		style.Render(content),
	)
}

func (r *Renderer) RenderSideBySide(diffContent string, width int) string {
	if diffContent == "" {
		return r.styles.Dim.Render("No diff content")
	}

	parsed := Parse(diffContent)
	var result strings.Builder

	halfWidth := (width - 3) / 2

	for i, file := range parsed.Files {
		if i > 0 {
			result.WriteString("\n")
		}

		result.WriteString(r.renderFileHeader(file))
		result.WriteString("\n")

		for _, hunk := range file.Hunks {
			result.WriteString(r.renderHunkHeader(hunk))
			result.WriteString("\n")

			oldLines := make([]Line, 0)
			newLines := make([]Line, 0)

			for _, line := range hunk.Lines {
				switch line.Type {
				case DiffRemove:
					oldLines = append(oldLines, line)
				case DiffAdd:
					newLines = append(newLines, line)
				case DiffContext:
					for len(oldLines) < len(newLines) {
						oldLines = append(oldLines, Line{Type: DiffContext, Content: ""})
					}
					for len(newLines) < len(oldLines) {
						newLines = append(newLines, Line{Type: DiffContext, Content: ""})
					}
					oldLines = append(oldLines, line)
					newLines = append(newLines, line)
				}
			}

			for len(oldLines) < len(newLines) {
				oldLines = append(oldLines, Line{Type: DiffContext, Content: ""})
			}
			for len(newLines) < len(oldLines) {
				newLines = append(newLines, Line{Type: DiffContext, Content: ""})
			}

			for i := 0; i < len(oldLines); i++ {
				left := r.renderSideLine(oldLines[i], halfWidth, true)
				right := r.renderSideLine(newLines[i], halfWidth, false)
				result.WriteString(fmt.Sprintf("%s │ %s\n", left, right))
			}
		}
	}

	return result.String()
}

func (r *Renderer) renderSideLine(line Line, width int, isLeft bool) string {
	style := r.styles.DiffContext

	switch line.Type {
	case DiffAdd:
		style = r.styles.DiffAdd
	case DiffRemove:
		style = r.styles.DiffRemove
	}

	content := line.Content
	if len(content) > width-6 && width > 6 {
		content = content[:width-9] + "..."
	}

	var lineNum string
	if isLeft && line.OldNumber > 0 {
		lineNum = fmt.Sprintf("%4d", line.OldNumber)
	} else if !isLeft && line.NewNumber > 0 {
		lineNum = fmt.Sprintf("%4d", line.NewNumber)
	} else {
		lineNum = "    "
	}

	padding := width - len(lineNum) - 1 - len(content)
	if padding < 0 {
		padding = 0
	}

	return fmt.Sprintf("%s %s%s",
		r.styles.LineNumber.Render(lineNum),
		style.Render(content),
		strings.Repeat(" ", padding),
	)
}
