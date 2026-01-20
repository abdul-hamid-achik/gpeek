package git

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	gogitconfig "github.com/go-git/go-git/v5/config"
)

// ConfigEntry represents a single git config entry
type ConfigEntry struct {
	Key      string
	Value    string
	Section  string
	Editable bool
}

// ConfigSection represents a group of config entries
type ConfigSection struct {
	Name    string
	Entries []ConfigEntry
}

// editableKeys lists config keys that can be safely edited
var editableKeys = map[string]bool{
	"user.name":           true,
	"user.email":          true,
	"core.editor":         true,
	"core.autocrlf":       true,
	"core.ignorecase":     true,
	"core.pager":          true,
	"core.whitespace":     true,
	"init.defaultBranch":  true,
	"pull.rebase":         true,
	"pull.ff":             true,
	"push.default":        true,
	"push.autoSetupRemote": true,
	"merge.conflictstyle": true,
	"merge.ff":            true,
	"diff.tool":           true,
	"merge.tool":          true,
	"fetch.prune":         true,
	"rebase.autoStash":    true,
	"commit.gpgsign":      true,
}

// GetConfig reads the local repository configuration
func (r *Repository) GetConfig() ([]ConfigSection, error) {
	cfg, err := r.repo.Config()
	if err != nil {
		return nil, err
	}

	var sections []ConfigSection

	// User section
	userSection := ConfigSection{Name: "User"}
	if cfg.User.Name != "" {
		userSection.Entries = append(userSection.Entries, ConfigEntry{
			Key:      "user.name",
			Value:    cfg.User.Name,
			Section:  "user",
			Editable: true,
		})
	}
	if cfg.User.Email != "" {
		userSection.Entries = append(userSection.Entries, ConfigEntry{
			Key:      "user.email",
			Value:    cfg.User.Email,
			Section:  "user",
			Editable: true,
		})
	}
	if len(userSection.Entries) > 0 {
		sections = append(sections, userSection)
	}

	// Core section
	coreSection := ConfigSection{Name: "Core"}
	if cfg.Core.IsBare {
		coreSection.Entries = append(coreSection.Entries, ConfigEntry{
			Key:      "core.bare",
			Value:    "true",
			Section:  "core",
			Editable: false,
		})
	}

	// Read raw config for additional core values
	rawCfg := cfg.Raw
	if rawCfg != nil {
		coreRaw := rawCfg.Section("core")
		if coreRaw != nil {
			for _, opt := range coreRaw.Options {
				key := "core." + opt.Key
				// Skip if already added
				if key == "core.bare" && cfg.Core.IsBare {
					continue
				}
				coreSection.Entries = append(coreSection.Entries, ConfigEntry{
					Key:      key,
					Value:    opt.Value,
					Section:  "core",
					Editable: editableKeys[key],
				})
			}
		}
	}
	if len(coreSection.Entries) > 0 {
		sections = append(sections, coreSection)
	}

	// Remotes section
	remotesSection := ConfigSection{Name: "Remotes"}
	for name, remote := range cfg.Remotes {
		urls := strings.Join(remote.URLs, ", ")
		remotesSection.Entries = append(remotesSection.Entries, ConfigEntry{
			Key:      "remote." + name + ".url",
			Value:    urls,
			Section:  "remote",
			Editable: false,
		})
	}
	if len(remotesSection.Entries) > 0 {
		sections = append(sections, remotesSection)
	}

	// Branches section
	branchesSection := ConfigSection{Name: "Branches"}
	for name, branch := range cfg.Branches {
		if branch.Remote != "" {
			branchesSection.Entries = append(branchesSection.Entries, ConfigEntry{
				Key:      "branch." + name + ".remote",
				Value:    branch.Remote,
				Section:  "branch",
				Editable: false,
			})
		}
		if branch.Merge.String() != "" {
			branchesSection.Entries = append(branchesSection.Entries, ConfigEntry{
				Key:      "branch." + name + ".merge",
				Value:    branch.Merge.Short(),
				Section:  "branch",
				Editable: false,
			})
		}
	}
	if len(branchesSection.Entries) > 0 {
		sections = append(sections, branchesSection)
	}

	return sections, nil
}

// GetConfigValue gets a specific config value
func (r *Repository) GetConfigValue(key string) (string, error) {
	cfg, err := r.repo.Config()
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid config key: %s", key)
	}

	section, name := parts[0], parts[1]

	switch section {
	case "user":
		switch name {
		case "name":
			return cfg.User.Name, nil
		case "email":
			return cfg.User.Email, nil
		}
	}

	// Try raw config for other values
	if cfg.Raw != nil {
		sec := cfg.Raw.Section(section)
		if sec != nil {
			return sec.Option(name), nil
		}
	}

	return "", fmt.Errorf("config key not found: %s", key)
}

// SetConfigValue sets a config value (local repository only)
func (r *Repository) SetConfigValue(key, value string) error {
	if !editableKeys[key] {
		return fmt.Errorf("config key %s is not editable", key)
	}

	// Validate the value
	if err := validateConfigValue(key, value); err != nil {
		return err
	}

	cfg, err := r.repo.Config()
	if err != nil {
		return err
	}

	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid config key: %s", key)
	}

	section, name := parts[0], parts[1]

	switch section {
	case "user":
		switch name {
		case "name":
			cfg.User.Name = value
		case "email":
			cfg.User.Email = value
		default:
			return r.setRawConfigValue(cfg, section, name, value)
		}
	default:
		return r.setRawConfigValue(cfg, section, name, value)
	}

	return r.repo.SetConfig(cfg)
}

// setRawConfigValue sets a value in the raw config
func (r *Repository) setRawConfigValue(cfg *gogitconfig.Config, section, name, value string) error {
	if cfg.Raw == nil {
		return fmt.Errorf("raw config not available")
	}

	sec := cfg.Raw.Section(section)
	sec.SetOption(name, value)

	return r.repo.SetConfig(cfg)
}

// validateConfigValue validates a config value based on its key
func validateConfigValue(key, value string) error {
	switch key {
	case "user.name":
		if strings.TrimSpace(value) == "" {
			return errors.New("user.name cannot be empty")
		}
	case "user.email":
		if strings.TrimSpace(value) == "" {
			return errors.New("user.email cannot be empty")
		}
		// Simple email validation
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(value) {
			return fmt.Errorf("invalid email format: %s", value)
		}
	case "core.autocrlf":
		validValues := map[string]bool{"true": true, "false": true, "input": true}
		if !validValues[strings.ToLower(value)] {
			return fmt.Errorf("core.autocrlf must be 'true', 'false', or 'input'")
		}
	case "core.ignorecase", "fetch.prune", "rebase.autoStash", "commit.gpgsign", "push.autoSetupRemote":
		validValues := map[string]bool{"true": true, "false": true}
		if !validValues[strings.ToLower(value)] {
			return fmt.Errorf("%s must be 'true' or 'false'", key)
		}
	case "pull.rebase":
		validValues := map[string]bool{"true": true, "false": true, "merges": true, "interactive": true}
		if !validValues[strings.ToLower(value)] {
			return fmt.Errorf("pull.rebase must be 'true', 'false', 'merges', or 'interactive'")
		}
	case "pull.ff", "merge.ff":
		validValues := map[string]bool{"true": true, "false": true, "only": true}
		if !validValues[strings.ToLower(value)] {
			return fmt.Errorf("%s must be 'true', 'false', or 'only'", key)
		}
	case "push.default":
		validValues := map[string]bool{"nothing": true, "current": true, "upstream": true, "simple": true, "matching": true}
		if !validValues[strings.ToLower(value)] {
			return fmt.Errorf("push.default must be 'nothing', 'current', 'upstream', 'simple', or 'matching'")
		}
	case "merge.conflictstyle":
		validValues := map[string]bool{"merge": true, "diff3": true, "zdiff3": true}
		if !validValues[strings.ToLower(value)] {
			return fmt.Errorf("merge.conflictstyle must be 'merge', 'diff3', or 'zdiff3'")
		}
	}

	return nil
}

// IsConfigEditable returns true if the config key is editable
func IsConfigEditable(key string) bool {
	return editableKeys[key]
}
