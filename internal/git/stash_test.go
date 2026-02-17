package git

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// parseStashLine replicates the parsing logic from StashList for testing.
// This allows testing the parsing without running git.
func parseStashLine(line string) (Stash, bool) {
	stashPattern := regexp.MustCompile(`stash@\{(\d+)\}`)

	parts := strings.SplitN(line, "|", 4)
	if len(parts) < 4 {
		return Stash{}, false
	}

	stashRef := parts[0]
	message := parts[1]
	hash := parts[2]
	timeStr := parts[3]

	matches := stashPattern.FindStringSubmatch(stashRef)
	if len(matches) < 2 {
		return Stash{}, false
	}
	index, _ := strconv.Atoi(matches[1])

	branch := ""
	if strings.HasPrefix(message, "WIP on ") || strings.HasPrefix(message, "On ") {
		colonIdx := strings.Index(message, ":")
		if colonIdx > 0 {
			prefix := message[:colonIdx]
			if strings.HasPrefix(prefix, "WIP on ") {
				branch = strings.TrimPrefix(prefix, "WIP on ")
			} else if strings.HasPrefix(prefix, "On ") {
				branch = strings.TrimPrefix(prefix, "On ")
			}
		}
	}

	parsedTime, _ := time.Parse("2006-01-02 15:04:05 -0700", timeStr)

	return Stash{
		Index:   index,
		Message: message,
		Branch:  branch,
		Hash:    hash,
		Time:    parsedTime,
	}, true
}

func TestParseStashLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantOk      bool
		wantIndex   int
		wantMessage string
		wantBranch  string
		wantHash    string
	}{
		{
			name:        "WIP on main",
			line:        "stash@{0}|WIP on main: abc1234 some commit message|deadbeefcafe1234567890abcdef1234567890ab|2025-01-15 10:30:00 +0000",
			wantOk:      true,
			wantIndex:   0,
			wantMessage: "WIP on main: abc1234 some commit message",
			wantBranch:  "main",
			wantHash:    "deadbeefcafe1234567890abcdef1234567890ab",
		},
		{
			name:        "On feature branch",
			line:        "stash@{1}|On feature/auth: my stash message|1234567890abcdef1234567890abcdef12345678|2025-02-01 14:00:00 -0500",
			wantOk:      true,
			wantIndex:   1,
			wantMessage: "On feature/auth: my stash message",
			wantBranch:  "feature/auth",
			wantHash:    "1234567890abcdef1234567890abcdef12345678",
		},
		{
			name:        "custom message no branch detection",
			line:        "stash@{2}|my custom stash message|abcdef1234567890abcdef1234567890abcdef12|2025-03-10 08:00:00 +0300",
			wantOk:      true,
			wantIndex:   2,
			wantMessage: "my custom stash message",
			wantBranch:  "",
			wantHash:    "abcdef1234567890abcdef1234567890abcdef12",
		},
		{
			name:        "high index",
			line:        "stash@{99}|WIP on develop: fix build|aabbccdd11223344556677889900aabbccddeeff|2025-01-01 00:00:00 +0000",
			wantOk:      true,
			wantIndex:   99,
			wantBranch:  "develop",
		},
		{
			name:   "too few parts",
			line:   "stash@{0}|message only",
			wantOk: false,
		},
		{
			name:   "no stash ref",
			line:   "invalid|message|hash|2025-01-01 00:00:00 +0000",
			wantOk: false,
		},
		{
			name:   "empty line",
			line:   "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stash, ok := parseStashLine(tt.line)
			if ok != tt.wantOk {
				t.Fatalf("parseStashLine ok = %v, want %v", ok, tt.wantOk)
			}
			if !ok {
				return
			}
			if stash.Index != tt.wantIndex {
				t.Errorf("index = %d, want %d", stash.Index, tt.wantIndex)
			}
			if tt.wantMessage != "" && stash.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", stash.Message, tt.wantMessage)
			}
			if stash.Branch != tt.wantBranch {
				t.Errorf("branch = %q, want %q", stash.Branch, tt.wantBranch)
			}
			if tt.wantHash != "" && stash.Hash != tt.wantHash {
				t.Errorf("hash = %q, want %q", stash.Hash, tt.wantHash)
			}
		})
	}
}

func TestParseStashBranchExtraction(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		wantBranch string
	}{
		{"WIP on main", "WIP on main: abc1234 commit msg", "main"},
		{"WIP on feature branch", "WIP on feature/login: def5678 add login", "feature/login"},
		{"On main", "On main: manual stash message", "main"},
		{"On branch with dashes", "On my-feature-branch: save progress", "my-feature-branch"},
		{"custom message", "my custom message without branch", ""},
		{"empty message", "", ""},
		{"WIP on with no colon", "WIP on something", ""},
		{"On with no colon", "On something", ""},
		{"starts with On but not prefix", "Only a test message", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate branch extraction logic
			branch := ""
			if strings.HasPrefix(tt.message, "WIP on ") || strings.HasPrefix(tt.message, "On ") {
				colonIdx := strings.Index(tt.message, ":")
				if colonIdx > 0 {
					prefix := tt.message[:colonIdx]
					if strings.HasPrefix(prefix, "WIP on ") {
						branch = strings.TrimPrefix(prefix, "WIP on ")
					} else if strings.HasPrefix(prefix, "On ") {
						branch = strings.TrimPrefix(prefix, "On ")
					}
				}
			}
			if branch != tt.wantBranch {
				t.Errorf("branch = %q, want %q", branch, tt.wantBranch)
			}
		})
	}
}

func TestParseStashTimeParsing(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string
		wantYear int
		wantZero bool
	}{
		{
			name:     "valid time",
			timeStr:  "2025-06-15 14:30:00 +0000",
			wantYear: 2025,
		},
		{
			name:     "valid time with timezone",
			timeStr:  "2024-12-25 08:00:00 -0500",
			wantYear: 2024,
		},
		{
			name:     "invalid time",
			timeStr:  "not-a-time",
			wantZero: true,
		},
		{
			name:     "empty time",
			timeStr:  "",
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, _ := time.Parse("2006-01-02 15:04:05 -0700", tt.timeStr)
			if tt.wantZero {
				if !parsed.IsZero() {
					t.Errorf("expected zero time, got %v", parsed)
				}
			} else {
				if parsed.Year() != tt.wantYear {
					t.Errorf("year = %d, want %d", parsed.Year(), tt.wantYear)
				}
			}
		})
	}
}

func TestStashStruct(t *testing.T) {
	s := Stash{
		Index:   3,
		Message: "WIP on main: fix bug",
		Branch:  "main",
		Hash:    "abc123",
		Time:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	if s.Index != 3 {
		t.Errorf("Index = %d, want 3", s.Index)
	}
	if s.Branch != "main" {
		t.Errorf("Branch = %q, want 'main'", s.Branch)
	}
}

func TestParseStashMultipleLines(t *testing.T) {
	// Simulate parsing multiple lines like StashList does
	output := `stash@{0}|WIP on main: abc1234 first stash|aaaa|2025-01-01 00:00:00 +0000
stash@{1}|On develop: second stash|bbbb|2025-01-02 00:00:00 +0000
stash@{2}|custom third stash|cccc|2025-01-03 00:00:00 +0000`

	lines := strings.Split(output, "\n")
	var stashes []Stash
	for _, line := range lines {
		if s, ok := parseStashLine(line); ok {
			stashes = append(stashes, s)
		}
	}

	if len(stashes) != 3 {
		t.Fatalf("expected 3 stashes, got %d", len(stashes))
	}

	if stashes[0].Branch != "main" {
		t.Errorf("stash 0 branch = %q, want 'main'", stashes[0].Branch)
	}
	if stashes[1].Branch != "develop" {
		t.Errorf("stash 1 branch = %q, want 'develop'", stashes[1].Branch)
	}
	if stashes[2].Branch != "" {
		t.Errorf("stash 2 branch = %q, want empty", stashes[2].Branch)
	}

	// Verify indices are correct
	for i, s := range stashes {
		if s.Index != i {
			t.Errorf("stash %d index = %d, want %d", i, s.Index, i)
		}
	}
}
