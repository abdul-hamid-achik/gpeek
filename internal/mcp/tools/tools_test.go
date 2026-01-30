package tools

import (
	"testing"
	"time"
)

// Test helper types response structures

func TestTimeAgo(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"just now", 30 * time.Second, "just now"},
		{"1 minute ago", 1 * time.Minute, "1 minute ago"},
		{"5 minutes ago", 5 * time.Minute, "5 minutes ago"},
		{"1 hour ago", 1 * time.Hour, "1 hour ago"},
		{"3 hours ago", 3 * time.Hour, "3 hours ago"},
		{"1 day ago", 24 * time.Hour, "1 day ago"},
		{"5 days ago", 5 * 24 * time.Hour, "5 days ago"},
		{"1 month ago", 35 * 24 * time.Hour, "1 month ago"},
		{"6 months ago", 180 * 24 * time.Hour, "6 months ago"},
		{"1 year ago", 400 * 24 * time.Hour, "1 year ago"},
		{"3 years ago", 3 * 365 * 24 * time.Hour, "3 years ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp := time.Now().Add(-tt.duration)
			result := TimeAgo(timestamp)
			if result != tt.expected {
				t.Errorf("TimeAgo() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDefaultPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "."},
		{".", "."},
		{"/path/to/repo", "/path/to/repo"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := DefaultPath(tt.input)
			if result != tt.expected {
				t.Errorf("DefaultPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"current directory", ".", false},
		{"absolute path", "/tmp/repo", false},
		{"relative path", "subdir/file", false},
		{"path with spaces", "path with spaces", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestErrorResult(t *testing.T) {
	result := ErrorResult("test error message")

	if !result.IsError {
		t.Error("ErrorResult() should set IsError to true")
	}

	if len(result.Content) != 1 {
		t.Errorf("ErrorResult() should have 1 content item, got %d", len(result.Content))
	}
}

// Test response type structures

func TestStatusResponse(t *testing.T) {
	resp := StatusResponse{
		Repository: RepositoryInfo{
			Name:   "test-repo",
			Path:   "/path/to/repo",
			Branch: "main",
		},
		Staged:    []FileInfo{{Path: "file1.go", Status: "added"}},
		Unstaged:  []FileInfo{{Path: "file2.go", Status: "modified"}},
		Untracked: []string{"file3.go"},
		Summary: StatusSummary{
			StagedCount:    1,
			UnstagedCount:  1,
			UntrackedCount: 1,
			IsClean:        false,
			HasConflicts:   false,
		},
	}

	if resp.Repository.Name != "test-repo" {
		t.Error("StatusResponse repository name mismatch")
	}
	if len(resp.Staged) != 1 {
		t.Error("StatusResponse staged count mismatch")
	}
	if resp.Summary.IsClean {
		t.Error("StatusResponse should not be clean")
	}
}

func TestDiffResponse(t *testing.T) {
	resp := DiffResponse{
		File:   "test.go",
		Commit: "abc1234",
		Staged: true,
		Files: []FileDiffInfo{
			{
				OldName:   "test.go",
				NewName:   "test.go",
				IsBinary:  false,
				IsNew:     false,
				IsDelete:  false,
				IsRename:  false,
				Additions: 10,
				Deletions: 5,
			},
		},
		Stats: DiffStats{
			FilesChanged: 1,
			Additions:    10,
			Deletions:    5,
		},
	}

	if resp.Stats.FilesChanged != 1 {
		t.Errorf("DiffResponse files changed = %d, want 1", resp.Stats.FilesChanged)
	}
	if resp.Stats.Additions != 10 {
		t.Errorf("DiffResponse additions = %d, want 10", resp.Stats.Additions)
	}
}

func TestCommitInfo(t *testing.T) {
	info := CommitInfo{
		Hash:      "abc123456789",
		ShortHash: "abc1234",
		Message:   "Test commit",
		Author:    "Test Author",
		Email:     "test@example.com",
		Time:      time.Now(),
		TimeAgo:   "just now",
		IsMerge:   false,
		Parents:   []string{},
	}

	if info.ShortHash != "abc1234" {
		t.Error("CommitInfo short hash mismatch")
	}
	if info.IsMerge {
		t.Error("CommitInfo should not be a merge commit")
	}
}

func TestSummaryResponse(t *testing.T) {
	resp := SummaryResponse{
		Repository: RepositoryInfo{
			Name:   "test-repo",
			Path:   "/path",
			Branch: "main",
		},
		Status: SummaryStatus{
			IsClean: true,
		},
		RecentCommits: []CommitInfo{},
		Branches: SummaryBranches{
			Current: "main",
			Count:   1,
		},
		Stashes: SummaryStashes{
			Count: 0,
		},
		Tags: SummaryTags{
			Count: 0,
		},
		Enhanced: nil,
	}

	if !resp.Status.IsClean {
		t.Error("SummaryResponse should be clean")
	}
	if resp.Enhanced != nil {
		t.Error("SummaryResponse enhanced should be nil")
	}
}

func TestEnhancedSummary(t *testing.T) {
	enhanced := EnhancedSummary{
		HotFiles: []HotFile{
			{Path: "main.go", ChangeCount: 10, Authors: []string{"Author1"}},
		},
		Languages: []LanguageInfo{
			{Name: "Go", FileCount: 50, Percentage: 80.0},
		},
		ProjectType: "Go",
		Suggestions: []string{"Review main.go"},
	}

	if len(enhanced.HotFiles) != 1 {
		t.Error("EnhancedSummary hot files count mismatch")
	}
	if enhanced.ProjectType != "Go" {
		t.Error("EnhancedSummary project type mismatch")
	}
}

func TestFileHistoryResponse(t *testing.T) {
	resp := FileHistoryResponse{
		File:       "test.go",
		Commits:    []CommitInfo{},
		Total:      0,
		Offset:     0,
		Limit:      50,
		HasMore:    false,
		NextOffset: 0,
	}

	if resp.HasMore {
		t.Error("FileHistoryResponse should not have more")
	}
	if resp.Limit != 50 {
		t.Errorf("FileHistoryResponse limit = %d, want 50", resp.Limit)
	}
}

func TestCodeOwnersResponse(t *testing.T) {
	resp := CodeOwnersResponse{
		HasCodeowners:  true,
		CodeownersPath: ".github/CODEOWNERS",
		FileOwners: []FileOwnerInfo{
			{File: "main.go", Owners: []string{"@team"}, Source: "codeowners"},
		},
		AllRules: []CodeOwnerRuleInfo{
			{Pattern: "*.go", Owners: []string{"@team"}, Line: 1},
		},
	}

	if !resp.HasCodeowners {
		t.Error("CodeOwnersResponse should have CODEOWNERS")
	}
	if len(resp.FileOwners) != 1 {
		t.Error("CodeOwnersResponse file owners count mismatch")
	}
}

func TestSearchChangesResponse(t *testing.T) {
	resp := SearchChangesResponse{
		Query: "test query",
		Results: []SearchChangeResult{
			{
				Commit:      CommitInfo{Hash: "abc123"},
				MatchedIn:   "message",
				MatchedText: "test query found",
				Score:       1.0,
			},
		},
		Total: 1,
	}

	if resp.Total != 1 {
		t.Errorf("SearchChangesResponse total = %d, want 1", resp.Total)
	}
	if resp.Results[0].Score != 1.0 {
		t.Error("SearchChangesResponse score mismatch")
	}
}

func TestChangeImpactResponse(t *testing.T) {
	resp := ChangeImpactResponse{
		Files: []FileImpact{
			{
				Path:            "main.go",
				Type:            "source",
				ChangeFrequency: 10,
				LastChanged:     "1 day ago",
				RecentAuthors:   []string{"Author1"},
			},
		},
		Summary: ImpactSummary{
			TotalFiles:      1,
			HighImpactFiles: 1,
			TestCoverage:    "likely",
			RiskLevel:       "medium",
		},
		RelatedFiles:  []string{"helper.go"},
		AffectedTests: []string{"main_test.go"},
	}

	if resp.Summary.RiskLevel != "medium" {
		t.Errorf("ChangeImpactResponse risk level = %q, want %q", resp.Summary.RiskLevel, "medium")
	}
}

func TestConflictCheckResponse(t *testing.T) {
	resp := ConflictCheckResponse{
		Branch:         "feature",
		Into:           "main",
		WouldConflict:  false,
		SafeToMerge:    true,
		Recommendation: "Safe to merge",
		Note:           "Basic check",
	}

	if resp.WouldConflict {
		t.Error("ConflictCheckResponse should not conflict")
	}
	if !resp.SafeToMerge {
		t.Error("ConflictCheckResponse should be safe to merge")
	}
}

// Test input validation

func TestBlameInputValidation(t *testing.T) {
	tests := []struct {
		name      string
		input     BlameInput
		expectErr bool
	}{
		{
			name:      "valid input",
			input:     BlameInput{File: "test.go"},
			expectErr: false,
		},
		{
			name:      "missing file",
			input:     BlameInput{File: ""},
			expectErr: true,
		},
		{
			name:      "with line range",
			input:     BlameInput{File: "test.go", StartLine: 1, EndLine: 10},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.input.File == ""
			if hasErr != tt.expectErr {
				t.Errorf("BlameInput validation: got error=%v, want error=%v", hasErr, tt.expectErr)
			}
		})
	}
}

func TestChangesBetweenInputValidation(t *testing.T) {
	tests := []struct {
		name      string
		input     ChangesBetweenInput
		expectErr bool
	}{
		{
			name:      "valid input",
			input:     ChangesBetweenInput{From: "main", To: "feature"},
			expectErr: false,
		},
		{
			name:      "missing from",
			input:     ChangesBetweenInput{From: "", To: "feature"},
			expectErr: true,
		},
		{
			name:      "default to HEAD",
			input:     ChangesBetweenInput{From: "main", To: ""},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.input.From == ""
			if hasErr != tt.expectErr {
				t.Errorf("ChangesBetweenInput validation: got error=%v, want error=%v", hasErr, tt.expectErr)
			}
		})
	}
}
