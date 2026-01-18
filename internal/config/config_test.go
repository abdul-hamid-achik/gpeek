package config

import (
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Theme != "nord" {
		t.Errorf("expected theme 'nord', got '%s'", cfg.Theme)
	}

	if !cfg.UI.ShowIcons {
		t.Error("expected ShowIcons to be true")
	}

	if !cfg.UI.RelativeDates {
		t.Error("expected RelativeDates to be true")
	}

	if !cfg.UI.ConfirmDestructive {
		t.Error("expected ConfirmDestructive to be true")
	}

	if len(cfg.Keys.Quit) != 2 {
		t.Errorf("expected 2 quit keys, got %d", len(cfg.Keys.Quit))
	}

	if cfg.Keys.Quit[0] != "q" {
		t.Errorf("expected first quit key 'q', got '%s'", cfg.Keys.Quit[0])
	}
}

func TestKeyConfigDefaults(t *testing.T) {
	cfg := Default()

	tests := []struct {
		name     string
		keys     []string
		expected string
	}{
		{"Help", cfg.Keys.Help, "?"},
		{"Stage", cfg.Keys.Stage, "s"},
		{"Unstage", cfg.Keys.Unstage, "u"},
		{"Commit", cfg.Keys.Commit, "c"},
		{"Push", cfg.Keys.Push, "P"},
		{"Pull", cfg.Keys.Pull, "p"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.keys) == 0 {
				t.Errorf("%s keys should not be empty", tt.name)
				return
			}
			if tt.keys[0] != tt.expected {
				t.Errorf("expected first %s key '%s', got '%s'", tt.name, tt.expected, tt.keys[0])
			}
		})
	}
}

func TestGitConfigDefaults(t *testing.T) {
	cfg := Default()

	if cfg.Git.AutoFetch {
		t.Error("expected AutoFetch to be false by default")
	}

	if cfg.Git.AutoFetchInterval != 300 {
		t.Errorf("expected AutoFetchInterval 300, got %d", cfg.Git.AutoFetchInterval)
	}

	if cfg.Git.SignCommits {
		t.Error("expected SignCommits to be false by default")
	}
}

func TestGitHubConfigDefaults(t *testing.T) {
	cfg := Default()

	if !cfg.GitHub.Enabled {
		t.Error("expected GitHub.Enabled to be true by default")
	}
}
