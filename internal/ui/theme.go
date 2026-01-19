package ui

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

type Theme struct {
	Name       string     `yaml:"name"`
	Background string     `yaml:"background"`
	Foreground string     `yaml:"foreground"`
	Primary    string     `yaml:"primary"`
	Secondary  string     `yaml:"secondary"`
	Accent     string     `yaml:"accent"`
	Muted      string     `yaml:"muted"`
	Subtle     string     `yaml:"subtle"`
	Border     string     `yaml:"border"`
	Selection  string     `yaml:"selection"`
	Added      string     `yaml:"added"`
	Removed    string     `yaml:"removed"`
	Modified   string     `yaml:"modified"`
	Renamed    string     `yaml:"renamed"`
	Untracked  string     `yaml:"untracked"`
	Conflict   string     `yaml:"conflict"`
	Error      string     `yaml:"error"`
	Warning    string     `yaml:"warning"`
	Success    string     `yaml:"success"`
	Info       string     `yaml:"info"`
	Syntax     SyntaxColors `yaml:"syntax"`
}

type SyntaxColors struct {
	Keyword   string `yaml:"keyword"`
	String    string `yaml:"string"`
	Number    string `yaml:"number"`
	Comment   string `yaml:"comment"`
	Function  string `yaml:"function"`
	Type      string `yaml:"type"`
	Variable  string `yaml:"variable"`
	Constant  string `yaml:"constant"`
	Operator  string `yaml:"operator"`
}

func NordTheme() Theme {
	return Theme{
		Name:       "nord",
		Background: "#2E3440",
		Foreground: "#ECEFF4",
		Primary:    "#88C0D0",
		Secondary:  "#81A1C1",
		Accent:     "#5E81AC",
		Muted:      "#D8DEE9", // nord4 - readable secondary text (was #4C566A)
		Subtle:     "#616E88", // For line numbers, less prominent but readable
		Border:     "#3B4252",
		Selection:  "#434C5E",
		Added:      "#A3BE8C",
		Removed:    "#BF616A",
		Modified:   "#EBCB8B",
		Renamed:    "#B48EAD",
		Untracked:  "#D08770",
		Conflict:   "#BF616A",
		Error:      "#BF616A",
		Warning:    "#EBCB8B",
		Success:    "#A3BE8C",
		Info:       "#88C0D0",
		Syntax: SyntaxColors{
			Keyword:   "#81A1C1",
			String:    "#A3BE8C",
			Number:    "#B48EAD",
			Comment:   "#616E88",
			Function:  "#88C0D0",
			Type:      "#8FBCBB",
			Variable:  "#D8DEE9",
			Constant:  "#EBCB8B",
			Operator:  "#81A1C1",
		},
	}
}

func CatppuccinMochaTheme() Theme {
	return Theme{
		Name:       "catppuccin-mocha",
		Background: "#1E1E2E",
		Foreground: "#CDD6F4",
		Primary:    "#89B4FA",
		Secondary:  "#74C7EC",
		Accent:     "#B4BEFE",
		Muted:      "#A6ADC8", // subtext0 - readable secondary text
		Subtle:     "#6C7086", // overlay0 - for line numbers
		Border:     "#313244",
		Selection:  "#45475A",
		Added:      "#A6E3A1",
		Removed:    "#F38BA8",
		Modified:   "#F9E2AF",
		Renamed:    "#CBA6F7",
		Untracked:  "#FAB387",
		Conflict:   "#F38BA8",
		Error:      "#F38BA8",
		Warning:    "#F9E2AF",
		Success:    "#A6E3A1",
		Info:       "#89B4FA",
		Syntax: SyntaxColors{
			Keyword:   "#CBA6F7",
			String:    "#A6E3A1",
			Number:    "#FAB387",
			Comment:   "#6C7086",
			Function:  "#89B4FA",
			Type:      "#F9E2AF",
			Variable:  "#CDD6F4",
			Constant:  "#FAB387",
			Operator:  "#89DCEB",
		},
	}
}

func GruvboxDarkTheme() Theme {
	return Theme{
		Name:       "gruvbox-dark",
		Background: "#282828",
		Foreground: "#EBDBB2",
		Primary:    "#83A598",
		Secondary:  "#458588",
		Accent:     "#D3869B",
		Muted:      "#BDAE93", // fg3 - readable secondary text
		Subtle:     "#928374", // gray - for line numbers
		Border:     "#3C3836",
		Selection:  "#504945",
		Added:      "#B8BB26",
		Removed:    "#FB4934",
		Modified:   "#FABD2F",
		Renamed:    "#D3869B",
		Untracked:  "#FE8019",
		Conflict:   "#FB4934",
		Error:      "#FB4934",
		Warning:    "#FABD2F",
		Success:    "#B8BB26",
		Info:       "#83A598",
		Syntax: SyntaxColors{
			Keyword:   "#FB4934",
			String:    "#B8BB26",
			Number:    "#D3869B",
			Comment:   "#928374",
			Function:  "#8EC07C",
			Type:      "#FABD2F",
			Variable:  "#EBDBB2",
			Constant:  "#D3869B",
			Operator:  "#FE8019",
		},
	}
}

func DraculaTheme() Theme {
	return Theme{
		Name:       "dracula",
		Background: "#282A36",
		Foreground: "#F8F8F2",
		Primary:    "#BD93F9",
		Secondary:  "#8BE9FD",
		Accent:     "#FF79C6",
		Muted:      "#BFBFBF", // readable secondary text
		Subtle:     "#6272A4", // comment color - for line numbers
		Border:     "#44475A",
		Selection:  "#44475A",
		Added:      "#50FA7B",
		Removed:    "#FF5555",
		Modified:   "#F1FA8C",
		Renamed:    "#FFB86C",
		Untracked:  "#FFB86C",
		Conflict:   "#FF5555",
		Error:      "#FF5555",
		Warning:    "#F1FA8C",
		Success:    "#50FA7B",
		Info:       "#8BE9FD",
		Syntax: SyntaxColors{
			Keyword:   "#FF79C6",
			String:    "#F1FA8C",
			Number:    "#BD93F9",
			Comment:   "#6272A4",
			Function:  "#50FA7B",
			Type:      "#8BE9FD",
			Variable:  "#F8F8F2",
			Constant:  "#BD93F9",
			Operator:  "#FF79C6",
		},
	}
}

func LoadTheme(name string) (Theme, error) {
	switch name {
	case "nord":
		return NordTheme(), nil
	case "catppuccin-mocha", "catppuccin":
		return CatppuccinMochaTheme(), nil
	case "gruvbox-dark", "gruvbox":
		return GruvboxDarkTheme(), nil
	case "dracula":
		return DraculaTheme(), nil
	}

	paths := []string{
		filepath.Join("themes", name+".yaml"),
	}

	if configDir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(configDir, "gpeek", "themes", name+".yaml"))
	}

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			var theme Theme
			if err := yaml.Unmarshal(data, &theme); err != nil {
				return Theme{}, err
			}
			return theme, nil
		}
	}

	return NordTheme(), nil
}

func (t Theme) Color(hex string) lipgloss.Color {
	return lipgloss.Color(hex)
}

func (t Theme) AdaptiveColor(light, dark string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Light: light, Dark: dark}
}
