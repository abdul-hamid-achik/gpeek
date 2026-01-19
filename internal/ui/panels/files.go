package panels

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type FileEntry struct {
	Path    string
	Status  git.FileStatus
	Staged  bool
	Section int
}

type FilesPanel struct {
	BasePanel
	styles   *ui.Styles
	viewport viewport.Model

	stagedFiles   []FileEntry
	unstagedFiles []FileEntry

	cursor   int
	section  int
	selected map[string]bool
}

func NewFilesPanel(styles *ui.Styles) *FilesPanel {
	vp := viewport.New(0, 0)
	return &FilesPanel{
		styles:   styles,
		viewport: vp,
		selected: make(map[string]bool),
	}
}

func (p *FilesPanel) SetStatus(status *git.Status) {
	p.stagedFiles = nil
	p.unstagedFiles = nil
	p.selected = make(map[string]bool)

	if status == nil {
		return
	}

	for _, f := range status.Staged {
		p.stagedFiles = append(p.stagedFiles, FileEntry{
			Path:    f.Path,
			Status:  f.Status,
			Staged:  true,
			Section: 0,
		})
	}

	for _, f := range status.Unstaged {
		p.unstagedFiles = append(p.unstagedFiles, FileEntry{
			Path:    f.Path,
			Status:  f.Status,
			Staged:  false,
			Section: 1,
		})
	}

	for _, f := range status.Untracked {
		p.unstagedFiles = append(p.unstagedFiles, FileEntry{
			Path:   f,
			Status: git.StatusUntracked,
			Staged: false,
			Section: 1,
		})
	}

	if p.cursor >= p.totalItems() && p.totalItems() > 0 {
		p.cursor = p.totalItems() - 1
	}
}

func (p *FilesPanel) Update(msg tea.Msg) tea.Cmd {
	if !p.focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			p.moveDown()
		case "k", "up":
			p.moveUp()
		case "g":
			p.cursor = 0
			p.section = 0
		case "G":
			if p.totalItems() > 0 {
				p.cursor = p.totalItems() - 1
				p.updateSection()
			}
		case "ctrl+d":
			for i := 0; i < p.height/2; i++ {
				p.moveDown()
			}
		case "ctrl+u":
			for i := 0; i < p.height/2; i++ {
				p.moveUp()
			}
		case " ":
			if file := p.SelectedFile(); file != nil {
				if p.selected[file.Path] {
					delete(p.selected, file.Path)
				} else {
					p.selected[file.Path] = true
				}
			}
		}
	}

	return nil
}

func (p *FilesPanel) View() string {
	if len(p.stagedFiles) == 0 && len(p.unstagedFiles) == 0 {
		return p.styles.Dim.Render("No changes\n\nWorking directory is clean")
	}

	var lines []string
	idx := 0

	if len(p.stagedFiles) > 0 {
		header := p.styles.Bold.Render(fmt.Sprintf("Staged (%d)", len(p.stagedFiles)))
		lines = append(lines, header)

		for _, f := range p.stagedFiles {
			line := p.renderFileEntry(f, idx == p.cursor)
			lines = append(lines, line)
			idx++
		}
	}

	if len(p.unstagedFiles) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		header := p.styles.Bold.Render(fmt.Sprintf("Unstaged (%d)", len(p.unstagedFiles)))
		lines = append(lines, header)

		for _, f := range p.unstagedFiles {
			line := p.renderFileEntry(f, idx == p.cursor)
			lines = append(lines, line)
			idx++
		}
	}

	content := strings.Join(lines, "\n")

	if len(lines) > p.height {
		start := 0
		cursorLine := p.getCursorLineIndex()
		if cursorLine > p.height-3 {
			start = cursorLine - p.height + 3
		}
		end := start + p.height
		if end > len(lines) {
			end = len(lines)
			start = end - p.height
			if start < 0 {
				start = 0
			}
		}
		content = strings.Join(lines[start:end], "\n")
	}

	return content
}

func (p *FilesPanel) renderFileEntry(f FileEntry, cursorOn bool) string {
	icon := p.statusIcon(f.Status, f.Staged)
	styleFunc := p.getStatusStyle(f.Status)

	prefix := "  "
	if cursorOn && p.focused {
		prefix = "> "
	}

	selectMark := "[ ]"
	if p.selected[f.Path] {
		selectMark = "[x]"
	}

	line := fmt.Sprintf("%s%s %s %s", prefix, selectMark, icon, f.Path)
	if cursorOn && p.focused {
		return p.styles.ListItemSelected.Render(line)
	}
	return styleFunc(line)
}

func (p *FilesPanel) statusIcon(status git.FileStatus, staged bool) string {
	if staged {
		return "✓"
	}
	switch status {
	case git.StatusModified:
		return "○"
	case git.StatusAdded:
		return "+"
	case git.StatusDeleted:
		return "-"
	case git.StatusRenamed:
		return "→"
	case git.StatusCopied:
		return "©"
	case git.StatusUntracked:
		return "?"
	default:
		return " "
	}
}

func (p *FilesPanel) getStatusStyle(status git.FileStatus) func(...string) string {
	switch status {
	case git.StatusAdded:
		return p.styles.Added.Render
	case git.StatusModified:
		return p.styles.Modified.Render
	case git.StatusDeleted:
		return p.styles.Removed.Render
	case git.StatusRenamed:
		return p.styles.Renamed.Render
	case git.StatusUntracked:
		return p.styles.Untracked.Render
	default:
		return p.styles.ListItem.Render
	}
}

func (p *FilesPanel) moveDown() {
	if p.cursor < p.totalItems()-1 {
		p.cursor++
		p.updateSection()
	}
}

func (p *FilesPanel) moveUp() {
	if p.cursor > 0 {
		p.cursor--
		p.updateSection()
	}
}

func (p *FilesPanel) totalItems() int {
	return len(p.stagedFiles) + len(p.unstagedFiles)
}

func (p *FilesPanel) updateSection() {
	if p.cursor < len(p.stagedFiles) {
		p.section = 0
	} else {
		p.section = 1
	}
}

func (p *FilesPanel) getCursorLineIndex() int {
	idx := 0
	if len(p.stagedFiles) > 0 {
		idx++
		if p.cursor < len(p.stagedFiles) {
			return idx + p.cursor
		}
		idx += len(p.stagedFiles)
	}
	if len(p.unstagedFiles) > 0 {
		if len(p.stagedFiles) > 0 {
			idx++
		}
		idx++
		return idx + (p.cursor - len(p.stagedFiles))
	}
	return p.cursor
}

func (p *FilesPanel) SelectedFile() *FileEntry {
	if p.totalItems() == 0 {
		return nil
	}
	if p.cursor < len(p.stagedFiles) {
		return &p.stagedFiles[p.cursor]
	}
	idx := p.cursor - len(p.stagedFiles)
	if idx < len(p.unstagedFiles) {
		return &p.unstagedFiles[idx]
	}
	return nil
}

func (p *FilesPanel) StagedFiles() []FileEntry {
	return p.stagedFiles
}

func (p *FilesPanel) SelectedFiles() []FileEntry {
	var files []FileEntry
	for _, f := range p.stagedFiles {
		if p.selected[f.Path] {
			files = append(files, f)
		}
	}
	for _, f := range p.unstagedFiles {
		if p.selected[f.Path] {
			files = append(files, f)
		}
	}
	return files
}

func (p *FilesPanel) HasSelection() bool {
	return len(p.selected) > 0
}

func (p *FilesPanel) ClearSelection() {
	p.selected = make(map[string]bool)
}

func (p *FilesPanel) SetSize(width, height int) {
	p.BasePanel.SetSize(width, height)
	p.viewport.Width = width
	p.viewport.Height = height
}
