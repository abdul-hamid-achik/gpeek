package modals

import (
	"testing"

	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func TestWorktreeModal_Modes(t *testing.T) {
	styles := ui.NewStyles(ui.NordTheme())
	modal := NewWorktreeModal(styles, nil, nil)

	// Test initial mode is list
	if modal.mode != WorktreeModeList {
		t.Errorf("expected initial mode to be WorktreeModeList, got %v", modal.mode)
	}

	// Test pressing 'n' switches to create mode
	modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if modal.mode != WorktreeModeCreate {
		t.Errorf("expected mode to be WorktreeModeCreate after pressing 'n', got %v", modal.mode)
	}

	// Test pressing 'esc' returns to list mode
	modal.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if modal.mode != WorktreeModeList {
		t.Errorf("expected mode to be WorktreeModeList after pressing 'esc', got %v", modal.mode)
	}
}

func TestWorktreeModal_ConfirmDelete(t *testing.T) {
	styles := ui.NewStyles(ui.NordTheme())
	// Create modal with a mock worktree
	modal := NewWorktreeModal(styles, nil, nil)

	// Without worktrees, 'd' should do nothing
	result, _ := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	if result == nil {
		t.Error("expected modal to remain open when deleting with no worktrees")
	}
}

func TestWorktreeModal_View(t *testing.T) {
	styles := ui.NewStyles(ui.NordTheme())
	modal := NewWorktreeModal(styles, nil, nil)

	view := modal.View()
	if view == "" {
		t.Error("expected View() to return non-empty string")
	}
}

func TestWorktreeModal_CreateViewValidation(t *testing.T) {
	styles := ui.NewStyles(ui.NordTheme())
	modal := NewWorktreeModal(styles, nil, nil)

	// Switch to create mode
	modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})

	// Try to create without path - should set error
	modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if modal.err != "Path is required" {
		t.Errorf("expected error 'Path is required', got '%s'", modal.err)
	}
}

func TestWorktreeModal_ConfirmDeleteNavigation(t *testing.T) {
	styles := ui.NewStyles(ui.NordTheme())
	modal := NewWorktreeModal(styles, nil, nil)

	// Manually set to confirm delete mode to test navigation
	modal.mode = WorktreeModeConfirmDelete
	modal.confirmPath = "/test/path"
	modal.confirmFocused = 1 // Default to "No"

	// Test tab switches focus
	modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if modal.confirmFocused != 0 {
		t.Errorf("expected confirmFocused to be 0 after tab, got %d", modal.confirmFocused)
	}

	// Test esc cancels
	result, _ := modal.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if modal.mode != WorktreeModeList {
		t.Errorf("expected mode to be WorktreeModeList after esc, got %v", modal.mode)
	}
	if result == nil {
		t.Error("expected modal to remain open after cancel")
	}
}
