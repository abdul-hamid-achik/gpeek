package panels

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/search"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	uisearch "github.com/abdul-hamid-achik/gpeek/internal/ui/search"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// FileSelectedMsg is sent when the selected file changes in the Files panel
type FileSelectedMsg struct {
	Path string
}

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

	allStagedFiles   []FileEntry
	allUnstagedFiles []FileEntry
	stagedFiles      []FileEntry
	unstagedFiles    []FileEntry

	cursor   int
	section  int
	selected map[string]bool

	// Track previous selection for change detection
	prevSelectedPath string

	// Filter support
	filterBar *uisearch.FilterBar
}

func NewFilesPanel(styles *ui.Styles) *FilesPanel {
	vp := viewport.New(0, 0)
	return &FilesPanel{
		styles:    styles,
		viewport:  vp,
		selected:  make(map[string]bool),
		filterBar: uisearch.NewFilterBar(styles),
	}
}

func (p *FilesPanel) SetStatus(status *git.Status) {
	p.allStagedFiles = nil
	p.allUnstagedFiles = nil
	p.selected = make(map[string]bool)

	if status == nil {
		p.applyFilter()
		return
	}

	for _, f := range status.Staged {
		p.allStagedFiles = append(p.allStagedFiles, FileEntry{
			Path:    f.Path,
			Status:  f.Status,
			Staged:  true,
			Section: 0,
		})
	}

	for _, f := range status.Unstaged {
		p.allUnstagedFiles = append(p.allUnstagedFiles, FileEntry{
			Path:    f.Path,
			Status:  f.Status,
			Staged:  false,
			Section: 1,
		})
	}

	for _, f := range status.Untracked {
		p.allUnstagedFiles = append(p.allUnstagedFiles, FileEntry{
			Path:    f,
			Status:  git.StatusUntracked,
			Staged:  false,
			Section: 1,
		})
	}

	p.applyFilter()
}

func (p *FilesPanel) applyFilter() {
	query := p.filterBar.GetQuery()

	if query.Text == "" {
		p.stagedFiles = p.allStagedFiles
		p.unstagedFiles = p.allUnstagedFiles
	} else {
		p.stagedFiles = search.Filter(p.allStagedFiles, query, func(f FileEntry) string {
			return f.Path
		})
		p.unstagedFiles = search.Filter(p.allUnstagedFiles, query, func(f FileEntry) string {
			return f.Path
		})
	}

	// Update filter bar counts
	totalCount := len(p.allStagedFiles) + len(p.allUnstagedFiles)
	matchCount := len(p.stagedFiles) + len(p.unstagedFiles)
	p.filterBar.SetCounts(matchCount, totalCount)

	// Adjust cursor if needed
	if p.cursor >= p.totalItems() && p.totalItems() > 0 {
		p.cursor = p.totalItems() - 1
	}
	if p.totalItems() == 0 {
		p.cursor = 0
	}
	p.updateSection()
}

func (p *FilesPanel) Update(msg tea.Msg) tea.Cmd {
	if !p.focused {
		return nil
	}

	// If filter bar is active, handle its input first
	if p.filterBar.IsActive() {
		cmd := p.filterBar.Update(msg)
		p.applyFilter()
		return cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			p.filterBar.Activate()
			return nil
		case "esc":
			if p.filterBar.HasFilter() {
				p.filterBar.Deactivate()
				p.applyFilter()
				return nil
			}
		case "j", "down":
			p.moveDown()
			// Check if selection changed and emit message
			if file := p.SelectedFile(); file != nil && file.Path != p.prevSelectedPath {
				p.prevSelectedPath = file.Path
				return func() tea.Msg {
					return FileSelectedMsg{Path: file.Path}
				}
			}
		case "k", "up":
			p.moveUp()
			// Check if selection changed and emit message
			if file := p.SelectedFile(); file != nil && file.Path != p.prevSelectedPath {
				p.prevSelectedPath = file.Path
				return func() tea.Msg {
					return FileSelectedMsg{Path: file.Path}
				}
			}
		case "g":
			p.cursor = 0
			p.section = 0
			// Check if selection changed and emit message
			if file := p.SelectedFile(); file != nil && file.Path != p.prevSelectedPath {
				p.prevSelectedPath = file.Path
				return func() tea.Msg {
					return FileSelectedMsg{Path: file.Path}
				}
			}
		case "G":
			if p.totalItems() > 0 {
				p.cursor = p.totalItems() - 1
				p.updateSection()
				// Check if selection changed and emit message
				if file := p.SelectedFile(); file != nil && file.Path != p.prevSelectedPath {
					p.prevSelectedPath = file.Path
					return func() tea.Msg {
						return FileSelectedMsg{Path: file.Path}
					}
				}
			}
		case "ctrl+d":
			oldPath := ""
			if file := p.SelectedFile(); file != nil {
				oldPath = file.Path
			}
			for i := 0; i < p.height/2; i++ {
				p.moveDown()
			}
			// Check if selection changed and emit message
			if file := p.SelectedFile(); file != nil && file.Path != oldPath {
				p.prevSelectedPath = file.Path
				return func() tea.Msg {
					return FileSelectedMsg{Path: file.Path}
				}
			}
		case "ctrl+u":
			oldPath := ""
			if file := p.SelectedFile(); file != nil {
				oldPath = file.Path
			}
			for i := 0; i < p.height/2; i++ {
				p.moveUp()
			}
			// Check if selection changed and emit message
			if file := p.SelectedFile(); file != nil && file.Path != oldPath {
				p.prevSelectedPath = file.Path
				return func() tea.Msg {
					return FileSelectedMsg{Path: file.Path}
				}
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
	totalFiles := len(p.allStagedFiles) + len(p.allUnstagedFiles)
	filteredFiles := len(p.stagedFiles) + len(p.unstagedFiles)

	if totalFiles == 0 {
		content := p.styles.Dim.Render("No changes\n\nWorking directory is clean")
		if p.filterBar.IsActive() || p.filterBar.HasFilter() {
			content += "\n" + p.filterBar.View()
		}
		return content
	}

	if filteredFiles == 0 && p.filterBar.HasFilter() {
		content := p.styles.Dim.Render("No matching files")
		if p.filterBar.IsActive() || p.filterBar.HasFilter() {
			content += "\n" + p.filterBar.View()
		}
		return content
	}

	var lines []string
	idx := 0

	if len(p.stagedFiles) > 0 {
		var header string
		if p.filterBar.HasFilter() {
			header = p.styles.Bold.Render(fmt.Sprintf("Staged (%d/%d)", len(p.stagedFiles), len(p.allStagedFiles)))
		} else {
			header = p.styles.Bold.Render(fmt.Sprintf("Staged (%d)", len(p.stagedFiles)))
		}
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
		var header string
		if p.filterBar.HasFilter() {
			header = p.styles.Bold.Render(fmt.Sprintf("Unstaged (%d/%d)", len(p.unstagedFiles), len(p.allUnstagedFiles)))
		} else {
			header = p.styles.Bold.Render(fmt.Sprintf("Unstaged (%d)", len(p.unstagedFiles)))
		}
		lines = append(lines, header)

		for _, f := range p.unstagedFiles {
			line := p.renderFileEntry(f, idx == p.cursor)
			lines = append(lines, line)
			idx++
		}
	}

	// Calculate available height for content (reserve space for filter bar)
	contentHeight := p.height
	if p.filterBar.IsActive() || p.filterBar.HasFilter() {
		contentHeight -= p.filterBar.FilterHeight()
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	content := strings.Join(lines, "\n")

	if len(lines) > contentHeight {
		start := 0
		cursorLine := p.getCursorLineIndex()
		if cursorLine > contentHeight-3 {
			start = cursorLine - contentHeight + 3
		}
		end := start + contentHeight
		if end > len(lines) {
			end = len(lines)
			start = end - contentHeight
			if start < 0 {
				start = 0
			}
		}
		content = strings.Join(lines[start:end], "\n")
	}

	// Add filter bar at bottom if active
	if p.filterBar.IsActive() || p.filterBar.HasFilter() {
		content += "\n" + p.filterBar.View()
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
	return p.allStagedFiles
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
	p.filterBar.SetWidth(width)
}

// IsFiltering returns true if the filter bar is active
func (p *FilesPanel) IsFiltering() bool {
	return p.filterBar.IsActive()
}

// ClearFilter clears the current filter
func (p *FilesPanel) ClearFilter() {
	p.filterBar.Deactivate()
	p.applyFilter()
}
