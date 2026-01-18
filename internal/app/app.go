package app

import (
	"time"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/abdul-hamid-achik/gpeek/internal/ui/modals"
	"github.com/abdul-hamid-achik/gpeek/internal/ui/panels"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width  int
	height int
	ready  bool

	layout  *ui.Layout
	styles  *ui.Styles
	keys    KeyMap
	focused ui.FocusedPanel

	repo *git.Repository

	filesPanel    *panels.FilesPanel
	branchesPanel *panels.BranchesPanel
	commitsPanel  *panels.CommitsPanel
	previewPanel  *panels.PreviewPanel

	activeModal modals.Modal

	statusMessage string
	statusError   bool
	statusTime    time.Time
}

type statusClearMsg struct{}
type refreshMsg struct{}
type gitStatusMsg struct {
	status *git.Status
	err    error
}
type gitBranchesMsg struct {
	branches []git.Branch
	current  string
	err      error
}
type gitCommitsMsg struct {
	commits []git.Commit
	err     error
}
type gitDiffMsg struct {
	diff string
	err  error
}

func New(repoPath string) (*Model, error) {
	repo, err := git.Open(repoPath)
	if err != nil {
		return nil, err
	}

	theme := ui.NordTheme()
	styles := ui.NewStyles(theme)

	m := &Model{
		layout:  ui.NewLayout(80, 24),
		styles:  styles,
		keys:    DefaultKeyMap(),
		focused: ui.PanelFiles,
		repo:    repo,
	}

	m.filesPanel = panels.NewFilesPanel(styles)
	m.branchesPanel = panels.NewBranchesPanel(styles)
	m.commitsPanel = panels.NewCommitsPanel(styles)
	m.previewPanel = panels.NewPreviewPanel(styles)

	m.filesPanel.Focus()

	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshStatus(),
		m.refreshBranches(),
		m.refreshCommits(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout.SetSize(msg.Width, msg.Height)
		m.ready = true

		filesDim := m.layout.FilesDimensions()
		branchesDim := m.layout.BranchesDimensions()
		commitsDim := m.layout.CommitsDimensions()
		previewDim := m.layout.PreviewDimensions()

		m.filesPanel.SetSize(filesDim.InnerWidth, filesDim.InnerHeight)
		m.branchesPanel.SetSize(branchesDim.InnerWidth, branchesDim.InnerHeight)
		m.commitsPanel.SetSize(commitsDim.InnerWidth, commitsDim.InnerHeight)
		m.previewPanel.SetSize(previewDim.InnerWidth, previewDim.InnerHeight)

	case tea.KeyMsg:
		if m.activeModal != nil {
			newModal, cmd := m.activeModal.Update(msg)
			if newModal == nil {
				m.activeModal = nil
			} else {
				m.activeModal = newModal
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.activeModal = modals.NewHelpModal(m.styles, m.keys.FullHelp())
			return m, nil

		case key.Matches(msg, m.keys.Refresh):
			cmds = append(cmds,
				m.refreshStatus(),
				m.refreshBranches(),
				m.refreshCommits(),
			)

		case key.Matches(msg, m.keys.FocusNext):
			m.cycleFocus(true)
			cmds = append(cmds, m.updatePreview())

		case key.Matches(msg, m.keys.FocusPrev):
			m.cycleFocus(false)
			cmds = append(cmds, m.updatePreview())

		case key.Matches(msg, m.keys.FocusFiles):
			m.setFocus(ui.PanelFiles)
			cmds = append(cmds, m.updatePreview())

		case key.Matches(msg, m.keys.FocusBranches):
			m.setFocus(ui.PanelBranches)
			cmds = append(cmds, m.updatePreview())

		case key.Matches(msg, m.keys.FocusCommits):
			m.setFocus(ui.PanelCommits)
			cmds = append(cmds, m.updatePreview())

		case key.Matches(msg, m.keys.FocusPreview):
			m.setFocus(ui.PanelPreview)

		case key.Matches(msg, m.keys.Stage):
			if m.focused == ui.PanelFiles {
				if file := m.filesPanel.SelectedFile(); file != nil {
					if err := m.repo.Stage(file.Path); err != nil {
						m.setStatus(err.Error(), true)
					} else {
						cmds = append(cmds, m.refreshStatus())
					}
				}
			}

		case key.Matches(msg, m.keys.Unstage):
			if m.focused == ui.PanelFiles {
				if file := m.filesPanel.SelectedFile(); file != nil {
					if err := m.repo.Unstage(file.Path); err != nil {
						m.setStatus(err.Error(), true)
					} else {
						cmds = append(cmds, m.refreshStatus())
					}
				}
			}

		case key.Matches(msg, m.keys.Commit):
			staged := m.filesPanel.StagedFiles()
			if len(staged) > 0 {
				m.activeModal = modals.NewCommitModal(m.styles, staged, func(message string) tea.Cmd {
					return func() tea.Msg {
						if err := m.repo.Commit(message); err != nil {
							return gitStatusMsg{err: err}
						}
						return refreshMsg{}
					}
				})
			} else {
				m.setStatus("No staged changes to commit", true)
			}

		case key.Matches(msg, m.keys.Push):
			cmds = append(cmds, func() tea.Msg {
				if err := m.repo.Push(); err != nil {
					return gitStatusMsg{err: err}
				}
				return refreshMsg{}
			})

		case key.Matches(msg, m.keys.Pull):
			cmds = append(cmds, func() tea.Msg {
				if err := m.repo.Pull(); err != nil {
					return gitStatusMsg{err: err}
				}
				return refreshMsg{}
			})

		case key.Matches(msg, m.keys.Fetch):
			cmds = append(cmds, func() tea.Msg {
				if err := m.repo.Fetch(); err != nil {
					return gitStatusMsg{err: err}
				}
				return refreshMsg{}
			})

		case key.Matches(msg, m.keys.Checkout), key.Matches(msg, m.keys.ShowCommit):
			if m.focused == ui.PanelBranches {
				if branch := m.branchesPanel.SelectedBranch(); branch != nil {
					if err := m.repo.Checkout(branch.Name); err != nil {
						m.setStatus(err.Error(), true)
					} else {
						cmds = append(cmds, m.refreshBranches(), m.refreshStatus(), m.refreshCommits())
					}
				}
			} else if m.focused == ui.PanelCommits {
				if commit := m.commitsPanel.SelectedCommit(); commit != nil {
					diff, _ := m.repo.CommitDiff(commit.Hash)
					title := commit.Hash[:7] + " - " + commit.Message
					if len(title) > 60 {
						title = title[:57] + "..."
					}
					m.activeModal = modals.NewDiffModal(m.styles, title, diff, m.width-4, m.height-4)
				}
			}

		case key.Matches(msg, m.keys.DiffMode):
			if m.focused == ui.PanelFiles {
				if file := m.filesPanel.SelectedFile(); file != nil {
					diff, _ := m.repo.FileDiff(file.Path, file.Staged)
					m.activeModal = modals.NewDiffModal(m.styles, file.Path, diff, m.width-4, m.height-4)
				}
			}

		case key.Matches(msg, m.keys.Worktree):
			worktrees, _ := m.repo.ListWorktrees()
			m.activeModal = modals.NewWorktreeModal(m.styles, worktrees, m.repo)

		default:
			cmd := m.updateFocusedPanel(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			cmds = append(cmds, m.updatePreview())
		}

	case gitStatusMsg:
		if msg.err != nil {
			m.setStatus(msg.err.Error(), true)
		} else {
			m.filesPanel.SetStatus(msg.status)
		}

	case gitBranchesMsg:
		if msg.err != nil {
			m.setStatus(msg.err.Error(), true)
		} else {
			m.branchesPanel.SetBranches(msg.branches, msg.current)
		}

	case gitCommitsMsg:
		if msg.err != nil {
			m.setStatus(msg.err.Error(), true)
		} else {
			m.commitsPanel.SetCommits(msg.commits)
		}

	case gitDiffMsg:
		if msg.err != nil {
			m.previewPanel.SetContent("Error loading diff: " + msg.err.Error())
		} else {
			m.previewPanel.SetDiff(msg.diff)
		}

	case refreshMsg:
		cmds = append(cmds,
			m.refreshStatus(),
			m.refreshBranches(),
			m.refreshCommits(),
		)

	case statusClearMsg:
		if time.Since(m.statusTime) >= 5*time.Second {
			m.statusMessage = ""
			m.statusError = false
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.layout.IsTooSmall() {
		return m.layout.TooSmallMessage()
	}

	filesDim := m.layout.FilesDimensions()
	branchesDim := m.layout.BranchesDimensions()
	commitsDim := m.layout.CommitsDimensions()
	previewDim := m.layout.PreviewDimensions()

	filesView := ui.RenderBorder(
		m.filesPanel.View(),
		"Files",
		filesDim.Width,
		filesDim.Height,
		m.focused == ui.PanelFiles,
		m.styles,
	)

	branchesView := ui.RenderBorder(
		m.branchesPanel.View(),
		"Branches",
		branchesDim.Width,
		branchesDim.Height,
		m.focused == ui.PanelBranches,
		m.styles,
	)

	commitsView := ui.RenderBorder(
		m.commitsPanel.View(),
		"Commits",
		commitsDim.Width,
		commitsDim.Height,
		m.focused == ui.PanelCommits,
		m.styles,
	)

	previewView := ui.RenderBorder(
		m.previewPanel.View(),
		"Preview",
		previewDim.Width,
		previewDim.Height,
		m.focused == ui.PanelPreview,
		m.styles,
	)

	leftColumn := lipgloss.JoinVertical(lipgloss.Left, filesView, branchesView)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, commitsView, previewView)
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	statusBar := m.renderStatusBar()

	view := lipgloss.JoinVertical(lipgloss.Left, mainView, statusBar)

	if m.activeModal != nil {
		modalView := m.activeModal.View()
		view = m.overlayModal(view, modalView)
	}

	return view
}

func (m *Model) renderStatusBar() string {
	repoName := m.repo.Name()
	branch := m.repo.CurrentBranch()

	var statusParts []string
	statusParts = append(statusParts, m.styles.StatusBarKey.Render(repoName))
	statusParts = append(statusParts, m.styles.StatusBarValue.Render(" on "))
	statusParts = append(statusParts, m.styles.StatusBarKey.Render(branch))

	if m.statusMessage != "" {
		style := m.styles.StatusBarValue
		if m.statusError {
			style = m.styles.StatusBarError
		}
		statusParts = append(statusParts, " │ "+style.Render(m.statusMessage))
	}

	left := lipgloss.JoinHorizontal(lipgloss.Left, statusParts...)

	hints := m.styles.HelpKey.Render("?") + m.styles.HelpDesc.Render(" help  ") +
		m.styles.HelpKey.Render("q") + m.styles.HelpDesc.Render(" quit")

	width := m.width
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(hints)
	padding := width - leftWidth - rightWidth - 2

	if padding < 0 {
		padding = 0
	}

	return m.styles.StatusBar.Width(width).Render(
		left + lipgloss.NewStyle().Width(padding).Render("") + hints,
	)
}

func (m *Model) overlayModal(base, modal string) string {
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceBackground(lipgloss.Color(m.styles.Theme.Background)),
	)
}

func (m *Model) cycleFocus(forward bool) {
	m.blurCurrent()
	if forward {
		m.focused = m.focused.Next()
	} else {
		m.focused = m.focused.Prev()
	}
	m.focusCurrent()
}

func (m *Model) setFocus(panel ui.FocusedPanel) {
	m.blurCurrent()
	m.focused = panel
	m.focusCurrent()
}

func (m *Model) blurCurrent() {
	switch m.focused {
	case ui.PanelFiles:
		m.filesPanel.Blur()
	case ui.PanelBranches:
		m.branchesPanel.Blur()
	case ui.PanelCommits:
		m.commitsPanel.Blur()
	case ui.PanelPreview:
		m.previewPanel.Blur()
	}
}

func (m *Model) focusCurrent() {
	switch m.focused {
	case ui.PanelFiles:
		m.filesPanel.Focus()
	case ui.PanelBranches:
		m.branchesPanel.Focus()
	case ui.PanelCommits:
		m.commitsPanel.Focus()
	case ui.PanelPreview:
		m.previewPanel.Focus()
	}
}

func (m *Model) updateFocusedPanel(msg tea.Msg) tea.Cmd {
	switch m.focused {
	case ui.PanelFiles:
		return m.filesPanel.Update(msg)
	case ui.PanelBranches:
		return m.branchesPanel.Update(msg)
	case ui.PanelCommits:
		return m.commitsPanel.Update(msg)
	case ui.PanelPreview:
		return m.previewPanel.Update(msg)
	}
	return nil
}

func (m *Model) updatePreview() tea.Cmd {
	switch m.focused {
	case ui.PanelFiles:
		if file := m.filesPanel.SelectedFile(); file != nil {
			return func() tea.Msg {
				diff, err := m.repo.FileDiff(file.Path, file.Staged)
				return gitDiffMsg{diff: diff, err: err}
			}
		}
	case ui.PanelCommits:
		if commit := m.commitsPanel.SelectedCommit(); commit != nil {
			return func() tea.Msg {
				diff, err := m.repo.CommitDiff(commit.Hash)
				return gitDiffMsg{diff: diff, err: err}
			}
		}
	}
	return nil
}

func (m *Model) refreshStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := m.repo.Status()
		return gitStatusMsg{status: status, err: err}
	}
}

func (m *Model) refreshBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.repo.ListBranches()
		current := m.repo.CurrentBranch()
		return gitBranchesMsg{branches: branches, current: current, err: err}
	}
}

func (m *Model) refreshCommits() tea.Cmd {
	return func() tea.Msg {
		commits, err := m.repo.Log(100)
		return gitCommitsMsg{commits: commits, err: err}
	}
}

func (m *Model) setStatus(msg string, isError bool) {
	m.statusMessage = msg
	m.statusError = isError
	m.statusTime = time.Now()
}

