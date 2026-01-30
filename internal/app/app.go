package app

import (
	"fmt"
	"time"

	"github.com/abdul-hamid-achik/gpeek/internal/config"
	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/abdul-hamid-achik/gpeek/internal/ui/modals"
	"github.com/abdul-hamid-achik/gpeek/internal/ui/panels"
	uisearch "github.com/abdul-hamid-achik/gpeek/internal/ui/search"
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

	activeModal     modals.Modal
	searchModal     *uisearch.SearchModal
	paletteModal    *modals.PaletteModal

	statusMessage   string
	statusError     bool
	statusTime      time.Time
	operationStatus string

	// Cached data for search
	cachedBranches []git.Branch
	cachedCommits  []git.Commit
}

type statusClearMsg struct{}
type refreshMsg struct{}
type operationDoneMsg struct {
	success string
	err     error
}
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
type gitTagsMsg struct {
	tags []git.Tag
	err  error
}

func New(repoPath string) (*Model, error) {
	repo, err := git.Open(repoPath)
	if err != nil {
		return nil, err
	}

	// Load config and theme
	cfg, _ := config.Load()
	theme, _ := ui.LoadTheme(cfg.Theme)
	styles := ui.NewStyles(theme)

	// Apply config settings
	if cfg.Git.DefaultRemote != "" {
		repo.SetDefaultRemote(cfg.Git.DefaultRemote)
	}

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
		m.refreshTags(),
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
		// Handle search modal first (it's separate from activeModal)
		// Handle command palette first
		if m.paletteModal != nil {
			result, cmd := m.paletteModal.Update(msg)
			if result == nil {
				m.paletteModal = nil
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		if m.searchModal != nil {
			result, cmd := m.searchModal.Update(msg)
			if result == nil {
				m.searchModal = nil
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

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
				if m.filesPanel.HasSelection() {
					files := m.filesPanel.SelectedFiles()
					var errors []string
					for _, f := range files {
						if !f.Staged {
							if err := m.repo.Stage(f.Path); err != nil {
								errors = append(errors, f.Path)
							}
						}
					}
					m.filesPanel.ClearSelection()
					if len(errors) > 0 {
						m.setStatus("Failed to stage some files", true)
					} else {
						m.setStatus(fmt.Sprintf("Staged %d files", len(files)), false)
					}
					cmds = append(cmds, m.refreshStatus())
				} else if file := m.filesPanel.SelectedFile(); file != nil {
					if err := m.repo.Stage(file.Path); err != nil {
						m.setStatus(err.Error(), true)
					} else {
						cmds = append(cmds, m.refreshStatus())
					}
				}
			}

		case key.Matches(msg, m.keys.Unstage):
			if m.focused == ui.PanelFiles {
				if m.filesPanel.HasSelection() {
					files := m.filesPanel.SelectedFiles()
					var errors []string
					for _, f := range files {
						if f.Staged {
							if err := m.repo.Unstage(f.Path); err != nil {
								errors = append(errors, f.Path)
							}
						}
					}
					m.filesPanel.ClearSelection()
					if len(errors) > 0 {
						m.setStatus("Failed to unstage some files", true)
					} else {
						m.setStatus(fmt.Sprintf("Unstaged %d files", len(files)), false)
					}
					cmds = append(cmds, m.refreshStatus())
				} else if file := m.filesPanel.SelectedFile(); file != nil {
					if err := m.repo.Unstage(file.Path); err != nil {
						m.setStatus(err.Error(), true)
					} else {
						cmds = append(cmds, m.refreshStatus())
					}
				}
			}

		case key.Matches(msg, m.keys.Discard):
			if m.focused == ui.PanelFiles {
				if file := m.filesPanel.SelectedFile(); file != nil {
					filePath := file.Path
					m.activeModal = modals.NewConfirmModal(
						m.styles,
						"Discard Changes",
						"Discard all changes to "+filePath+"?\nThis cannot be undone.",
						func() tea.Cmd {
							return func() tea.Msg {
								if err := m.repo.Discard(filePath); err != nil {
									return operationDoneMsg{err: err}
								}
								return operationDoneMsg{success: "Discarded changes to " + filePath}
							}
						},
					)
				}
			}

		case key.Matches(msg, m.keys.Commit):
			staged := m.filesPanel.StagedFiles()
			if len(staged) > 0 {
				lastMsg, lastHash, _ := m.repo.GetLastCommitInfo()
				m.activeModal = modals.NewCommitModal(m.styles, staged, lastMsg, lastHash, func(message string, isAmend bool) tea.Cmd {
					return func() tea.Msg {
						var err error
						if isAmend {
							err = m.repo.AmendCommit(message)
						} else {
							err = m.repo.Commit(message)
						}
						if err != nil {
							return gitStatusMsg{err: err}
						}
						return refreshMsg{}
					}
				})
			} else {
				m.setStatus("No staged changes to commit", true)
			}

		case key.Matches(msg, m.keys.Push):
			branch := m.repo.CurrentBranch()
			m.activeModal = modals.NewConfirmModal(
				m.styles,
				"Push",
				"Push commits to remote for branch '"+branch+"'?",
				func() tea.Cmd {
					return func() tea.Msg {
						if err := m.repo.Push(); err != nil {
							return operationDoneMsg{err: err}
						}
						return operationDoneMsg{success: "Pushed to remote"}
					}
				},
			)

		case key.Matches(msg, m.keys.Pull):
			m.operationStatus = "Pulling..."
			cmds = append(cmds, func() tea.Msg {
				if err := m.repo.Pull(); err != nil {
					return operationDoneMsg{err: err}
				}
				return operationDoneMsg{success: "Pulled from remote"}
			})

		case key.Matches(msg, m.keys.Fetch):
			m.operationStatus = "Fetching..."
			cmds = append(cmds, func() tea.Msg {
				if err := m.repo.Fetch(); err != nil {
					return operationDoneMsg{err: err}
				}
				return operationDoneMsg{success: "Fetched from remote"}
			})

		case key.Matches(msg, m.keys.Checkout), key.Matches(msg, m.keys.ShowCommit):
			switch m.focused {
			case ui.PanelBranches:
				if branch := m.branchesPanel.SelectedBranch(); branch != nil {
					if err := m.repo.Checkout(branch.Name); err != nil {
						m.setStatus(err.Error(), true)
					} else {
						m.setStatus("Switched to "+branch.Name, false)
						cmds = append(cmds, m.refreshBranches(), m.refreshStatus(), m.refreshCommits())
					}
				}
			case ui.PanelCommits:
				if commit := m.commitsPanel.SelectedCommit(); commit != nil {
					diff, _ := m.repo.CommitDiff(commit.Hash)
					title := commit.Hash[:7] + " - " + commit.Message
					titleWidth := lipgloss.Width(title)
					maxWidth := m.width - 8
					if titleWidth > maxWidth {
						title = title[:maxWidth-3] + "..."
					}
					m.activeModal = modals.NewDiffModal(m.styles, title, diff, m.width-4, m.height-4)
				}
			}

		case key.Matches(msg, m.keys.NewBranch):
			if m.focused == ui.PanelBranches {
				m.activeModal = modals.NewInputModal(
					m.styles,
					"New Branch",
					"Enter branch name:",
					"feature/my-branch",
					func(name string) tea.Cmd {
						return func() tea.Msg {
							if err := m.repo.CreateBranch(name); err != nil {
								return operationDoneMsg{err: err}
							}
							return operationDoneMsg{success: "Created branch " + name}
						}
					},
				)
			}

		case key.Matches(msg, m.keys.DeleteBranch):
			if m.focused == ui.PanelBranches {
				if branch := m.branchesPanel.SelectedBranch(); branch != nil {
					if branch.Name == m.repo.CurrentBranch() {
						m.setStatus("Cannot delete current branch", true)
					} else {
						branchName := branch.Name
						m.activeModal = modals.NewConfirmModal(
							m.styles,
							"Delete Branch",
							"Delete branch '"+branchName+"'?\nThis cannot be undone.",
							func() tea.Cmd {
								return func() tea.Msg {
									if err := m.repo.DeleteBranch(branchName); err != nil {
										return operationDoneMsg{err: err}
									}
									return operationDoneMsg{success: "Deleted branch " + branchName}
								}
							},
						)
					}
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

		case key.Matches(msg, m.keys.Stash):
			stashes, _ := m.repo.StashList()
			m.activeModal = modals.NewStashModal(m.styles, stashes, m.repo, m.width-8, m.height-8)

		case key.Matches(msg, m.keys.Blame):
			if m.focused == ui.PanelFiles {
				if file := m.filesPanel.SelectedFile(); file != nil {
					lines, err := m.repo.BlameFile(file.Path)
					if err != nil {
						m.setStatus("Cannot blame file: "+err.Error(), true)
					} else {
						m.activeModal = modals.NewBlameModal(m.styles, file.Path, lines, m.repo, m.width-4, m.height-4)
					}
				}
			}

		case key.Matches(msg, m.keys.GlobalSearch):
			// Open global search modal
			worktrees, _ := m.repo.ListWorktrees()
			m.searchModal = uisearch.NewSearchModal(
				m.styles,
				m.cachedBranches,
				m.cachedCommits,
				worktrees,
				m.repo.CurrentBranch(),
				m.width-8,
				m.height-8,
			)
			m.searchModal.SetCallbacks(
				func(branch *git.Branch) tea.Cmd {
					return func() tea.Msg {
						if err := m.repo.Checkout(branch.Name); err != nil {
							return operationDoneMsg{err: err}
						}
						return operationDoneMsg{success: "Switched to " + branch.Name}
					}
				},
				func(commit *git.Commit) tea.Cmd {
					// Show commit diff
					diff, _ := m.repo.CommitDiff(commit.Hash)
					title := commit.Hash[:7] + " - " + commit.Message
					m.activeModal = modals.NewDiffModal(m.styles, title, diff, m.width-4, m.height-4)
					return nil
				},
				nil, // No worktree action for now
			)
			return m, nil

		case key.Matches(msg, m.keys.GitConfig):
			// Open git config modal
			m.activeModal = modals.NewGitConfigModal(m.styles, m.repo, m.width-8, m.height-8)
			return m, nil

		case key.Matches(msg, m.keys.CommandPalette):
			// Open command palette
			m.paletteModal = modals.NewPaletteModal(
				modals.DefaultCommands(),
				m.executeCommand,
				m.width-20,
				m.height-10,
			)
			return m, nil

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
			m.cachedBranches = msg.branches // Cache for search
		}

	case gitCommitsMsg:
		if msg.err != nil {
			m.setStatus(msg.err.Error(), true)
		} else {
			m.commitsPanel.SetCommits(msg.commits)
			m.cachedCommits = msg.commits // Cache for search
		}

	case gitDiffMsg:
		if msg.err != nil {
			m.previewPanel.SetContent("Error loading diff: " + msg.err.Error())
		} else {
			m.previewPanel.SetDiff(msg.diff)
		}

	case gitTagsMsg:
		if msg.err == nil {
			m.branchesPanel.SetTags(msg.tags)
		}

	case refreshMsg:
		cmds = append(cmds,
			m.refreshStatus(),
			m.refreshBranches(),
			m.refreshCommits(),
			m.refreshTags(),
		)

	case operationDoneMsg:
		m.operationStatus = ""
		if msg.err != nil {
			m.setStatus(msg.err.Error(), true)
		} else if msg.success != "" {
			m.setStatus(msg.success, false)
			cmds = append(cmds,
				m.refreshStatus(),
				m.refreshBranches(),
				m.refreshCommits(),
				m.refreshTags(),
			)
		}

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

	if m.searchModal != nil {
		modalView := m.searchModal.View()
		view = m.overlayModal(view, modalView)
	}

	if m.paletteModal != nil {
		modalView := m.paletteModal.View()
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

	if m.operationStatus != "" {
		statusParts = append(statusParts, " │ "+m.styles.Spinner.Render(m.operationStatus))
	} else if m.statusMessage != "" {
		style := m.styles.Success
		if m.statusError {
			style = m.styles.StatusBarError
		}
		statusParts = append(statusParts, " │ "+style.Render(m.statusMessage))
	}

	left := lipgloss.JoinHorizontal(lipgloss.Left, statusParts...)

	hints := m.getPanelHints()

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

func (m *Model) getPanelHints() string {
	switch m.focused {
	case ui.PanelFiles:
		return m.styles.HelpKey.Render("s") + m.styles.HelpDesc.Render(" stage  ") +
			m.styles.HelpKey.Render("u") + m.styles.HelpDesc.Render(" unstage  ") +
			m.styles.HelpKey.Render("b") + m.styles.HelpDesc.Render(" blame  ") +
			m.styles.HelpKey.Render("?") + m.styles.HelpDesc.Render(" help")
	case ui.PanelBranches:
		return m.styles.HelpKey.Render("enter") + m.styles.HelpDesc.Render(" checkout  ") +
			m.styles.HelpKey.Render("n") + m.styles.HelpDesc.Render(" new  ") +
			m.styles.HelpKey.Render("t") + m.styles.HelpDesc.Render(" tags  ") +
			m.styles.HelpKey.Render("?") + m.styles.HelpDesc.Render(" help")
	case ui.PanelCommits:
		return m.styles.HelpKey.Render("enter") + m.styles.HelpDesc.Render(" view diff  ") +
			m.styles.HelpKey.Render("?") + m.styles.HelpDesc.Render(" help")
	case ui.PanelPreview:
		return m.styles.HelpKey.Render("j/k") + m.styles.HelpDesc.Render(" scroll  ") +
			m.styles.HelpKey.Render("?") + m.styles.HelpDesc.Render(" help")
	default:
		return m.styles.HelpKey.Render("?") + m.styles.HelpDesc.Render(" help  ") +
			m.styles.HelpKey.Render("q") + m.styles.HelpDesc.Render(" quit")
	}
}

func (m *Model) overlayModal(_, modal string) string {
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceBackground(lipgloss.Color(m.styles.Theme.Background)),
	)
}

// executeCommand handles command palette selections
func (m *Model) executeCommand(cmd modals.Command) tea.Cmd {
	switch cmd.ID {
	case "focus_files":
		m.setFocus(ui.PanelFiles)
	case "focus_branches":
		m.setFocus(ui.PanelBranches)
	case "focus_commits":
		m.setFocus(ui.PanelCommits)
	case "focus_preview":
		m.setFocus(ui.PanelPreview)
	case "refresh":
		return tea.Batch(m.refreshStatus(), m.refreshBranches(), m.refreshCommits(), m.refreshTags())
	case "help":
		m.activeModal = modals.NewHelpModal(m.styles, m.keys.FullHelp())
	case "quit":
		return tea.Quit
	case "git_config":
		m.activeModal = modals.NewGitConfigModal(m.styles, m.repo, m.width-8, m.height-8)
	// Commands that need more context - just show status for now
	case "commit", "stash", "worktree", "stage", "unstage", "discard", "push", "pull", "fetch":
		m.setStatus(fmt.Sprintf("Use keyboard shortcut for '%s'", cmd.Title), false)
	default:
		m.setStatus(fmt.Sprintf("Command '%s' executed", cmd.Title), false)
	}
	return nil
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

func (m *Model) refreshTags() tea.Cmd {
	return func() tea.Msg {
		tags, err := m.repo.ListTags()
		return gitTagsMsg{tags: tags, err: err}
	}
}

func (m *Model) setStatus(msg string, isError bool) {
	m.statusMessage = msg
	m.statusError = isError
	m.statusTime = time.Now()
}

