package search

import (
	"fmt"
	"sort"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/search"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func truncHash(h string) string {
	if len(h) > 7 {
		return h[:7]
	}
	return h
}

// SearchResultType identifies the type of search result
type SearchResultType int

const (
	ResultTypeBranch SearchResultType = iota
	ResultTypeCommit
	ResultTypeWorktree
)

// SearchResult represents a single search result
type SearchResult struct {
	Type        SearchResultType
	Label       string // Display text
	Description string // Secondary info
	Score       int    // Match score for sorting
	Data        interface{}
}

// SearchModal provides global search across all data sources
type SearchModal struct {
	styles *ui.Styles
	input  textinput.Model
	width  int
	height int
	closed bool

	// Search options
	caseSensitive bool
	regexMode     bool

	// Data sources
	branches  []git.Branch
	commits   []git.Commit
	worktrees []git.Worktree
	current   string // Current branch name

	// Results
	results       []SearchResult
	selectedIndex int

	// Callbacks
	onSelectBranch   func(branch *git.Branch) tea.Cmd
	onSelectCommit   func(commit *git.Commit) tea.Cmd
	onSelectWorktree func(worktree *git.Worktree) tea.Cmd
}

// NewSearchModal creates a new global search modal
func NewSearchModal(
	styles *ui.Styles,
	branches []git.Branch,
	commits []git.Commit,
	worktrees []git.Worktree,
	current string,
	width, height int,
) *SearchModal {
	ti := textinput.New()
	ti.Placeholder = "Search branches, commits, worktrees..."
	ti.CharLimit = 100
	ti.Width = width - 20
	ti.Focus()

	m := &SearchModal{
		styles:        styles,
		input:         ti,
		width:         width,
		height:        height,
		branches:      branches,
		commits:       commits,
		worktrees:     worktrees,
		current:       current,
		selectedIndex: 0,
	}

	// Initialize with all results
	m.search("")
	return m
}

// SetCallbacks sets the callback functions for result selection
func (m *SearchModal) SetCallbacks(
	onBranch func(*git.Branch) tea.Cmd,
	onCommit func(*git.Commit) tea.Cmd,
	onWorktree func(*git.Worktree) tea.Cmd,
) {
	m.onSelectBranch = onBranch
	m.onSelectCommit = onCommit
	m.onSelectWorktree = onWorktree
}

// search performs the search and updates results
func (m *SearchModal) search(pattern string) {
	m.results = nil

	opts := search.QueryOptions{
		DefaultMode:      search.ModeFuzzy,
		DefaultCaseSens:  m.caseSensitive,
		DefaultSmartCase: !m.caseSensitive,
	}

	if m.regexMode {
		opts.DefaultMode = search.ModeRegex
	}

	query := search.ParseQuery(pattern, opts)
	matcher := query.CreateMatcher()

	// Search branches
	for i := range m.branches {
		b := &m.branches[i]
		result := matcher.Match(b.Name)
		if result.Matched || pattern == "" {
			label := b.Name
			desc := ""
			if b.Name == m.current {
				desc = "(current)"
			}
			m.results = append(m.results, SearchResult{
				Type:        ResultTypeBranch,
				Label:       label,
				Description: desc,
				Score:       result.Score,
				Data:        b,
			})
		}
	}

	// Search commits (by message and author)
	for i := range m.commits {
		c := &m.commits[i]
		msgResult := matcher.Match(c.Message)
		authorResult := matcher.Match(c.Author)

		if msgResult.Matched || authorResult.Matched || pattern == "" {
			score := msgResult.Score
			if authorResult.Score > score {
				score = authorResult.Score
			}

			// Limit commit message length
			msg := c.Message
			if len(msg) > 50 {
				msg = msg[:47] + "..."
			}

			m.results = append(m.results, SearchResult{
				Type:        ResultTypeCommit,
				Label:       truncHash(c.Hash) + " " + msg,
				Description: c.Author,
				Score:       score,
				Data:        c,
			})
		}
	}

	// Search worktrees
	for i := range m.worktrees {
		w := &m.worktrees[i]
		pathResult := matcher.Match(w.Path)
		branchResult := matcher.Match(w.Branch)

		if pathResult.Matched || branchResult.Matched || pattern == "" {
			score := pathResult.Score
			if branchResult.Score > score {
				score = branchResult.Score
			}

			m.results = append(m.results, SearchResult{
				Type:        ResultTypeWorktree,
				Label:       w.Path,
				Description: w.Branch,
				Score:       score,
				Data:        w,
			})
		}
	}

	// Sort results by score (highest first)
	sort.Slice(m.results, func(i, j int) bool {
		// Group by type first
		if m.results[i].Type != m.results[j].Type {
			return m.results[i].Type < m.results[j].Type
		}
		return m.results[i].Score > m.results[j].Score
	})

	// Reset selection
	m.selectedIndex = 0
}

// ShouldClose returns true if the modal should close
func (m *SearchModal) ShouldClose() bool {
	return m.closed
}

// Update handles input events - implements modals.Modal interface
func (m *SearchModal) Update(msg tea.Msg) (interface{}, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.closed = true
			return nil, nil
		case "enter":
			return nil, m.selectCurrent()
		case "down", "ctrl+n", "j":
			if len(m.results) > 0 {
				m.selectedIndex++
				if m.selectedIndex >= len(m.results) {
					m.selectedIndex = 0
				}
			}
			return m, nil
		case "up", "ctrl+p", "k":
			if len(m.results) > 0 {
				m.selectedIndex--
				if m.selectedIndex < 0 {
					m.selectedIndex = len(m.results) - 1
				}
			}
			return m, nil
		case "alt+c":
			m.caseSensitive = !m.caseSensitive
			m.search(m.input.Value())
			return m, nil
		case "alt+r":
			m.regexMode = !m.regexMode
			m.search(m.input.Value())
			return m, nil
		}
	}

	prevValue := m.input.Value()
	m.input, cmd = m.input.Update(msg)

	// If search text changed, re-search
	if m.input.Value() != prevValue {
		m.search(m.input.Value())
	}

	return m, cmd
}

// selectCurrent handles selection of the current result
func (m *SearchModal) selectCurrent() tea.Cmd {
	if len(m.results) == 0 || m.selectedIndex >= len(m.results) {
		m.closed = true
		return nil
	}

	result := m.results[m.selectedIndex]
	m.closed = true

	switch result.Type {
	case ResultTypeBranch:
		if m.onSelectBranch != nil {
			return m.onSelectBranch(result.Data.(*git.Branch))
		}
	case ResultTypeCommit:
		if m.onSelectCommit != nil {
			return m.onSelectCommit(result.Data.(*git.Commit))
		}
	case ResultTypeWorktree:
		if m.onSelectWorktree != nil {
			return m.onSelectWorktree(result.Data.(*git.Worktree))
		}
	}

	return nil
}

// View renders the search modal
func (m *SearchModal) View() string {
	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Primary)).
		Bold(true).
		Padding(0, 1)

	title := titleStyle.Render("Search")

	// Search input
	inputLine := "  > " + m.input.View()

	// Mode indicators
	var indicators []string
	if m.regexMode {
		indicators = append(indicators, m.styles.StatusBarKey.Render("[~]"))
	}
	if m.caseSensitive {
		indicators = append(indicators, m.styles.StatusBarKey.Render("[Aa]"))
	} else {
		indicators = append(indicators, m.styles.Dim.Render("[Aa]"))
	}

	inputLine += "  " + strings.Join(indicators, " ")

	// Results grouped by type
	var resultLines []string

	// Group results by type
	var branchResults, commitResults, worktreeResults []SearchResult
	for _, r := range m.results {
		switch r.Type {
		case ResultTypeBranch:
			branchResults = append(branchResults, r)
		case ResultTypeCommit:
			commitResults = append(commitResults, r)
		case ResultTypeWorktree:
			worktreeResults = append(worktreeResults, r)
		}
	}

	// Calculate the absolute index for rendering
	currentIdx := 0

	// Render branches
	if len(branchResults) > 0 {
		header := m.styles.Bold.Render(fmt.Sprintf("─ Branches (%d) ", len(branchResults)))
		header += m.styles.Dim.Render(strings.Repeat("─", m.width-lipgloss.Width(header)-6))
		resultLines = append(resultLines, header)

		maxBranches := 5
		for i, r := range branchResults {
			if i >= maxBranches {
				more := fmt.Sprintf("  ... and %d more", len(branchResults)-maxBranches)
				resultLines = append(resultLines, m.styles.Dim.Render(more))
				currentIdx += len(branchResults) - maxBranches
				break
			}
			resultLines = append(resultLines, m.renderResult(r, currentIdx == m.selectedIndex))
			currentIdx++
		}
		resultLines = append(resultLines, "")
	}

	// Render commits
	if len(commitResults) > 0 {
		header := m.styles.Bold.Render(fmt.Sprintf("─ Commits (%d) ", len(commitResults)))
		header += m.styles.Dim.Render(strings.Repeat("─", m.width-lipgloss.Width(header)-6))
		resultLines = append(resultLines, header)

		maxCommits := 5
		for i, r := range commitResults {
			if i >= maxCommits {
				more := fmt.Sprintf("  ... and %d more", len(commitResults)-maxCommits)
				resultLines = append(resultLines, m.styles.Dim.Render(more))
				currentIdx += len(commitResults) - maxCommits
				break
			}
			resultLines = append(resultLines, m.renderResult(r, currentIdx == m.selectedIndex))
			currentIdx++
		}
		resultLines = append(resultLines, "")
	}

	// Render worktrees
	if len(worktreeResults) > 0 {
		header := m.styles.Bold.Render(fmt.Sprintf("─ Worktrees (%d) ", len(worktreeResults)))
		header += m.styles.Dim.Render(strings.Repeat("─", m.width-lipgloss.Width(header)-6))
		resultLines = append(resultLines, header)

		maxWorktrees := 3
		for i, r := range worktreeResults {
			if i >= maxWorktrees {
				more := fmt.Sprintf("  ... and %d more", len(worktreeResults)-maxWorktrees)
				resultLines = append(resultLines, m.styles.Dim.Render(more))
				break
			}
			resultLines = append(resultLines, m.renderResult(r, currentIdx == m.selectedIndex))
			currentIdx++
		}
	}

	if len(m.results) == 0 && m.input.Value() != "" {
		resultLines = append(resultLines, m.styles.Dim.Render("  No results found"))
	}

	results := strings.Join(resultLines, "\n")

	// Footer
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Muted))

	footer := footerStyle.Render("Enter: select  j/k: navigate  Esc: close")

	// Compose body
	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		inputLine,
		"",
		results,
		"",
		footer,
	)

	// Apply modal style
	return m.styles.Modal.
		Width(m.width).
		Height(m.height).
		Render(body)
}

// renderResult renders a single search result
func (m *SearchModal) renderResult(r SearchResult, selected bool) string {
	prefix := "    "
	if selected {
		prefix = "  > "
	}

	var line string
	switch r.Type {
	case ResultTypeBranch:
		if r.Description != "" {
			line = fmt.Sprintf("%s%s  %s", prefix, r.Label, m.styles.Dim.Render(r.Description))
		} else {
			line = fmt.Sprintf("%s%s", prefix, r.Label)
		}
	case ResultTypeCommit:
		line = fmt.Sprintf("%s%s", prefix, r.Label)
	case ResultTypeWorktree:
		line = fmt.Sprintf("%s%s  %s", prefix, r.Label, m.styles.Dim.Render("("+r.Description+")"))
	}

	if selected {
		return m.styles.ListItemSelected.Render(line)
	}
	return m.styles.ListItem.Render(line)
}
