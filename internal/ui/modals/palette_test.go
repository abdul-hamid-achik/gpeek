package modals

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewPaletteModal(t *testing.T) {
	commands := DefaultCommands()
	modal := NewPaletteModal(commands, nil, 80, 24)

	if modal == nil {
		t.Fatal("NewPaletteModal returned nil")
	}

	if len(modal.commands) != len(commands) {
		t.Errorf("commands count = %d, want %d", len(modal.commands), len(commands))
	}

	if len(modal.filtered) != len(commands) {
		t.Errorf("filtered count = %d, want %d", len(modal.filtered), len(commands))
	}

	if modal.selected != 0 {
		t.Errorf("selected = %d, want 0", modal.selected)
	}
}

func TestPaletteModalFiltering(t *testing.T) {
	commands := DefaultCommands()
	modal := NewPaletteModal(commands, nil, 80, 24)

	// Initial state - all commands
	if len(modal.filtered) != len(commands) {
		t.Errorf("initial filtered = %d, want %d", len(modal.filtered), len(commands))
	}

	// Simulate typing "commit"
	modal.input.SetValue("commit")
	modal.filterCommands()

	// Should filter to commands with "commit" in title/description
	foundCommit := false
	for _, cmd := range modal.filtered {
		if strings.Contains(strings.ToLower(cmd.Title), "commit") ||
			strings.Contains(strings.ToLower(cmd.Description), "commit") {
			foundCommit = true
			break
		}
	}
	if !foundCommit && len(modal.filtered) > 0 {
		t.Error("filtering by 'commit' should include commit-related commands")
	}

	// Clear filter - should restore all commands
	modal.input.SetValue("")
	modal.filterCommands()

	if len(modal.filtered) != len(commands) {
		t.Errorf("after clear, filtered = %d, want %d", len(modal.filtered), len(commands))
	}
}

func TestPaletteModalCaseInsensitiveFilter(t *testing.T) {
	commands := DefaultCommands()
	modal := NewPaletteModal(commands, nil, 80, 24)

	// Filter with uppercase
	modal.input.SetValue("PUSH")
	modal.filterCommands()
	upperCount := len(modal.filtered)

	// Filter with lowercase
	modal.input.SetValue("push")
	modal.filterCommands()
	lowerCount := len(modal.filtered)

	if upperCount != lowerCount {
		t.Errorf("case sensitivity: upper=%d, lower=%d, should be equal", upperCount, lowerCount)
	}
}

func TestPaletteModalNavigation(t *testing.T) {
	commands := DefaultCommands()
	modal := NewPaletteModal(commands, nil, 80, 24)

	if modal.selected != 0 {
		t.Fatalf("initial selected = %d, want 0", modal.selected)
	}

	// Navigate down
	modal.Update(tea.KeyMsg{Type: tea.KeyDown})
	if modal.selected != 1 {
		t.Errorf("after down, selected = %d, want 1", modal.selected)
	}

	// Navigate up
	modal.Update(tea.KeyMsg{Type: tea.KeyUp})
	if modal.selected != 0 {
		t.Errorf("after up, selected = %d, want 0", modal.selected)
	}

	// Try to go above 0
	modal.Update(tea.KeyMsg{Type: tea.KeyUp})
	if modal.selected != 0 {
		t.Errorf("should not go below 0, selected = %d", modal.selected)
	}
}

func TestPaletteModalClose(t *testing.T) {
	commands := DefaultCommands()
	modal := NewPaletteModal(commands, nil, 80, 24)

	if modal.ShouldClose() {
		t.Error("modal should not be closed initially")
	}

	// Press Escape
	result, _ := modal.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if result != nil {
		t.Error("pressing Escape should return nil modal")
	}
}

func TestPaletteModalExecute(t *testing.T) {
	executed := false
	var executedCmd Command

	onExecute := func(cmd Command) tea.Cmd {
		executed = true
		executedCmd = cmd
		return nil
	}

	commands := []Command{
		{ID: "test", Title: "Test Command", Category: "Test"},
	}

	modal := NewPaletteModal(commands, onExecute, 80, 24)

	// Press Enter to execute
	result, _ := modal.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !executed {
		t.Error("command should have been executed")
	}

	if executedCmd.ID != "test" {
		t.Errorf("executed command ID = %q, want %q", executedCmd.ID, "test")
	}

	if result != nil {
		t.Error("modal should close after execution (return nil)")
	}
}

func TestPaletteModalView(t *testing.T) {
	commands := []Command{
		{ID: "cmd1", Title: "Command One", Category: "Cat1", Keybinding: "a"},
		{ID: "cmd2", Title: "Command Two", Category: "Cat2", Keybinding: "b"},
	}

	modal := NewPaletteModal(commands, nil, 80, 24)
	view := modal.View()

	// View should contain command titles
	if !strings.Contains(view, "Command One") {
		t.Error("view should contain 'Command One'")
	}
	if !strings.Contains(view, "Command Two") {
		t.Error("view should contain 'Command Two'")
	}

	// View should contain keybindings
	if !strings.Contains(view, "[a]") {
		t.Error("view should contain keybinding '[a]'")
	}
}

func TestPaletteModalEmptyResults(t *testing.T) {
	commands := DefaultCommands()
	modal := NewPaletteModal(commands, nil, 80, 24)

	// Filter with something that won't match
	modal.input.SetValue("xyznonexistent123")
	modal.filterCommands()

	if len(modal.filtered) != 0 {
		t.Errorf("should have 0 filtered results, got %d", len(modal.filtered))
	}

	view := modal.View()
	if !strings.Contains(view, "No commands found") {
		t.Error("view should show 'No commands found' message")
	}
}

func TestDefaultCommands(t *testing.T) {
	commands := DefaultCommands()

	if len(commands) == 0 {
		t.Fatal("DefaultCommands returned empty list")
	}

	// Check that all commands have required fields
	for i, cmd := range commands {
		if cmd.ID == "" {
			t.Errorf("command %d has empty ID", i)
		}
		if cmd.Title == "" {
			t.Errorf("command %d (%s) has empty Title", i, cmd.ID)
		}
		if cmd.Category == "" {
			t.Errorf("command %d (%s) has empty Category", i, cmd.ID)
		}
	}

	// Check for expected categories
	categories := make(map[string]bool)
	for _, cmd := range commands {
		categories[cmd.Category] = true
	}

	expectedCategories := []string{"Navigation", "Git", "Branch", "Search", "System"}
	for _, cat := range expectedCategories {
		if !categories[cat] {
			t.Errorf("missing expected category: %s", cat)
		}
	}
}

func TestPaletteStyles(t *testing.T) {
	styles := DefaultPaletteStyles()

	// Verify styles are initialized (basic smoke test)
	// Just check that the function returns without panic
	_ = styles.Container
	_ = styles.Input
	_ = styles.Item
	_ = styles.ItemSelected
	_ = styles.ItemTitle
}

func TestPaletteModalCategoryFiltering(t *testing.T) {
	commands := DefaultCommands()
	modal := NewPaletteModal(commands, nil, 80, 24)

	// Filter by category name
	modal.input.SetValue("git")
	modal.filterCommands()

	// Should find Git category commands
	if len(modal.filtered) == 0 {
		t.Error("filtering by 'git' should find commands")
	}

	for _, cmd := range modal.filtered {
		matchesTitle := strings.Contains(strings.ToLower(cmd.Title), "git")
		matchesDesc := strings.Contains(strings.ToLower(cmd.Description), "git")
		matchesCat := strings.Contains(strings.ToLower(cmd.Category), "git")
		if !matchesTitle && !matchesDesc && !matchesCat {
			t.Errorf("command %q doesn't match 'git' filter", cmd.ID)
		}
	}
}

func TestPaletteModalSelectionBounds(t *testing.T) {
	commands := []Command{
		{ID: "cmd1", Title: "One", Category: "Test"},
		{ID: "cmd2", Title: "Two", Category: "Test"},
		{ID: "cmd3", Title: "Three", Category: "Test"},
	}

	modal := NewPaletteModal(commands, nil, 80, 24)

	// Go to end
	modal.selected = len(commands) - 1

	// Try to go past end
	modal.Update(tea.KeyMsg{Type: tea.KeyDown})
	if modal.selected != len(commands)-1 {
		t.Errorf("should not go past end, selected = %d, max = %d", modal.selected, len(commands)-1)
	}
}

func TestPaletteModalJumpCommands(t *testing.T) {
	commands := DefaultCommands()
	modal := NewPaletteModal(commands, nil, 80, 24)

	// Jump to end with ctrl+d
	modal.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if modal.selected != len(modal.filtered)-1 {
		t.Errorf("ctrl+d should jump to end, selected = %d, want %d", modal.selected, len(modal.filtered)-1)
	}

	// Jump to start with ctrl+u
	modal.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	if modal.selected != 0 {
		t.Errorf("ctrl+u should jump to start, selected = %d, want 0", modal.selected)
	}
}
