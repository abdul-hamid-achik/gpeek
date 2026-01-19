package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Theme  string       `yaml:"theme"`
	UI     UIConfig     `yaml:"ui"`
	Keys   KeyConfig    `yaml:"keys"`
	Git    GitConfig    `yaml:"git"`
	GitHub GitHubConfig `yaml:"github"`
	Search SearchConfig `yaml:"search"`
}

type UIConfig struct {
	ShowIcons          bool   `yaml:"show_icons"`
	DateFormat         string `yaml:"date_format"`
	RelativeDates      bool   `yaml:"relative_dates"`
	ConfirmDestructive bool   `yaml:"confirm_destructive"`
	ShowHiddenFiles    bool   `yaml:"show_hidden_files"`
}

type KeyConfig struct {
	Quit          []string `yaml:"quit"`
	Help          []string `yaml:"help"`
	Refresh       []string `yaml:"refresh"`
	FocusNext     []string `yaml:"focus_next"`
	FocusPrev     []string `yaml:"focus_prev"`
	FocusFiles    []string `yaml:"focus_files"`
	FocusBranches []string `yaml:"focus_branches"`
	FocusCommits  []string `yaml:"focus_commits"`
	FocusPreview  []string `yaml:"focus_preview"`
	Up            []string `yaml:"up"`
	Down          []string `yaml:"down"`
	PageUp        []string `yaml:"page_up"`
	PageDown      []string `yaml:"page_down"`
	Top           []string `yaml:"top"`
	Bottom        []string `yaml:"bottom"`
	Select        []string `yaml:"select"`
	Confirm       []string `yaml:"confirm"`
	Cancel        []string `yaml:"cancel"`
	Stage         []string `yaml:"stage"`
	Unstage       []string `yaml:"unstage"`
	Discard       []string `yaml:"discard"`
	Commit        []string `yaml:"commit"`
	Push          []string `yaml:"push"`
	Pull          []string `yaml:"pull"`
	Fetch         []string `yaml:"fetch"`
	Checkout      []string `yaml:"checkout"`
	NewBranch     []string `yaml:"new_branch"`
	DeleteBranch  []string `yaml:"delete_branch"`
	DiffMode      []string `yaml:"diff_mode"`
	Worktree      []string `yaml:"worktree"`
}

type GitConfig struct {
	AutoFetch         bool `yaml:"auto_fetch"`
	AutoFetchInterval int  `yaml:"auto_fetch_interval"`
	SignCommits       bool `yaml:"sign_commits"`
}

type GitHubConfig struct {
	Enabled bool `yaml:"enabled"`
}

type SearchConfig struct {
	DefaultMode      string `yaml:"default_mode"`      // "fuzzy", "exact", "regex"
	CaseSensitive    bool   `yaml:"case_sensitive"`
	SmartCase        bool   `yaml:"smart_case"`        // Auto case-sensitive if uppercase
	MaxResults       int    `yaml:"max_results"`       // Default: 100
	DebounceMs       int    `yaml:"debounce_ms"`       // Default: 150
	HighlightMatches bool   `yaml:"highlight_matches"`
}

func Default() *Config {
	return &Config{
		Theme: "nord",
		UI: UIConfig{
			ShowIcons:          true,
			DateFormat:         "2006-01-02 15:04",
			RelativeDates:      true,
			ConfirmDestructive: true,
			ShowHiddenFiles:    false,
		},
		Keys: KeyConfig{
			Quit:          []string{"q", "ctrl+c"},
			Help:          []string{"?"},
			Refresh:       []string{"ctrl+r"},
			FocusNext:     []string{"tab"},
			FocusPrev:     []string{"shift+tab"},
			FocusFiles:    []string{"1"},
			FocusBranches: []string{"2"},
			FocusCommits:  []string{"3"},
			FocusPreview:  []string{"4"},
			Up:            []string{"k", "up"},
			Down:          []string{"j", "down"},
			PageUp:        []string{"ctrl+u"},
			PageDown:      []string{"ctrl+d"},
			Top:           []string{"g"},
			Bottom:        []string{"G"},
			Select:        []string{"space"},
			Confirm:       []string{"enter"},
			Cancel:        []string{"esc"},
			Stage:         []string{"s"},
			Unstage:       []string{"u"},
			Discard:       []string{"x"},
			Commit:        []string{"c"},
			Push:          []string{"P"},
			Pull:          []string{"p"},
			Fetch:         []string{"f"},
			Checkout:      []string{"enter"},
			NewBranch:     []string{"n"},
			DeleteBranch:  []string{"d"},
			DiffMode:      []string{"v"},
			Worktree:      []string{"w"},
		},
		Git: GitConfig{
			AutoFetch:         false,
			AutoFetchInterval: 300,
			SignCommits:       false,
		},
		GitHub: GitHubConfig{
			Enabled: true,
		},
		Search: SearchConfig{
			DefaultMode:      "fuzzy",
			CaseSensitive:    false,
			SmartCase:        true,
			MaxResults:       100,
			DebounceMs:       150,
			HighlightMatches: true,
		},
	}
}

func Load() (*Config, error) {
	config := Default()

	paths := []string{}

	if configDir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(configDir, "gpeek", "config.yaml"))
		paths = append(paths, filepath.Join(configDir, "gpeek", "config.yml"))
	}

	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "gpeek", "config.yaml"))
		paths = append(paths, filepath.Join(home, ".config", "gpeek", "config.yml"))
		paths = append(paths, filepath.Join(home, ".gpeek.yaml"))
		paths = append(paths, filepath.Join(home, ".gpeek.yml"))
	}

	paths = append(paths, "configs/default.yaml")
	paths = append(paths, ".gpeek.yaml")
	paths = append(paths, ".gpeek.yml")

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, err
			}
			break
		}
	}

	return config, nil
}

func (c *Config) Save() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	gpeekDir := filepath.Join(configDir, "gpeek")
	if err := os.MkdirAll(gpeekDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(gpeekDir, "config.yaml")

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func ConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "gpeek"), nil
}
