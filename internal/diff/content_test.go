package diff

import (
	"testing"
)

func TestCountFileChanges(t *testing.T) {
	tests := []struct {
		name     string
		file     FileDiff
		wantAdds int
		wantDels int
	}{
		{
			name:     "empty file",
			file:     FileDiff{},
			wantAdds: 0,
			wantDels: 0,
		},
		{
			name: "only additions",
			file: FileDiff{
				Hunks: []Hunk{{
					Lines: []Line{
						{Type: DiffAdd, Content: "a"},
						{Type: DiffAdd, Content: "b"},
						{Type: DiffAdd, Content: "c"},
					},
				}},
			},
			wantAdds: 3,
			wantDels: 0,
		},
		{
			name: "only removals",
			file: FileDiff{
				Hunks: []Hunk{{
					Lines: []Line{
						{Type: DiffRemove, Content: "x"},
						{Type: DiffRemove, Content: "y"},
					},
				}},
			},
			wantAdds: 0,
			wantDels: 2,
		},
		{
			name: "mixed with context",
			file: FileDiff{
				Hunks: []Hunk{{
					Lines: []Line{
						{Type: DiffContext, Content: "ctx"},
						{Type: DiffRemove, Content: "old"},
						{Type: DiffAdd, Content: "new1"},
						{Type: DiffAdd, Content: "new2"},
						{Type: DiffContext, Content: "ctx2"},
					},
				}},
			},
			wantAdds: 2,
			wantDels: 1,
		},
		{
			name: "multiple hunks",
			file: FileDiff{
				Hunks: []Hunk{
					{Lines: []Line{
						{Type: DiffAdd, Content: "a1"},
						{Type: DiffRemove, Content: "r1"},
					}},
					{Lines: []Line{
						{Type: DiffAdd, Content: "a2"},
						{Type: DiffAdd, Content: "a3"},
					}},
				},
			},
			wantAdds: 3,
			wantDels: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adds, dels := CountFileChanges(tt.file)
			if adds != tt.wantAdds {
				t.Errorf("adds = %d, want %d", adds, tt.wantAdds)
			}
			if dels != tt.wantDels {
				t.Errorf("dels = %d, want %d", dels, tt.wantDels)
			}
		})
	}
}

func TestIsAllCollapsed(t *testing.T) {
	tests := []struct {
		name     string
		expanded map[int]bool
		want     bool
	}{
		{
			name:     "nil map",
			expanded: nil,
			want:     true,
		},
		{
			name:     "empty map",
			expanded: map[int]bool{},
			want:     true,
		},
		{
			name:     "all false",
			expanded: map[int]bool{0: false, 1: false, 2: false},
			want:     true,
		},
		{
			name:     "one expanded",
			expanded: map[int]bool{0: false, 1: true, 2: false},
			want:     false,
		},
		{
			name:     "all expanded",
			expanded: map[int]bool{0: true, 1: true},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAllCollapsed(tt.expanded); got != tt.want {
				t.Errorf("IsAllCollapsed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderContentNilDiff(t *testing.T) {
	styles := ContentStyles{} // zero-value styles

	content, positions := RenderContent(nil, nil, 0, styles, nil, nil)
	if content != "No changes" {
		t.Errorf("expected 'No changes' for nil diff, got %q", content)
	}
	if positions != nil {
		t.Errorf("expected nil positions for nil diff, got %v", positions)
	}
}

func TestRenderContentEmptyDiff(t *testing.T) {
	styles := ContentStyles{}
	d := &Diff{Files: []FileDiff{}}

	content, positions := RenderContent(d, nil, 0, styles, nil, nil)
	if content != "No changes" {
		t.Errorf("expected 'No changes' for empty diff, got %q", content)
	}
	if positions != nil {
		t.Errorf("expected nil positions for empty diff, got %v", positions)
	}
}

func TestRenderContentCollapsedFiles(t *testing.T) {
	styles := ContentStyles{}
	d := &Diff{
		Files: []FileDiff{
			{
				OldName: "file1.go",
				NewName: "file1.go",
				Hunks: []Hunk{{
					Lines: []Line{
						{Type: DiffAdd, Content: "added"},
					},
				}},
			},
			{
				OldName: "file2.go",
				NewName: "file2.go",
			},
		},
	}

	// All collapsed (empty expanded map)
	expanded := map[int]bool{}
	_, positions := RenderContent(d, expanded, 0, styles, nil, nil)

	if len(positions) != 2 {
		t.Fatalf("expected 2 file positions, got %d", len(positions))
	}

	// Collapsed files should each occupy 1 line (header only)
	for i, pos := range positions {
		if pos.Expanded {
			t.Errorf("file %d should not be expanded", i)
		}
	}
}

func TestRenderContentExpandedFile(t *testing.T) {
	styles := ContentStyles{}
	d := &Diff{
		Files: []FileDiff{
			{
				OldName: "file.go",
				NewName: "file.go",
				Hunks: []Hunk{{
					Header: "@@ -1,2 +1,3 @@",
					Lines: []Line{
						{Type: DiffContext, Content: "ctx"},
						{Type: DiffAdd, Content: "new"},
						{Type: DiffContext, Content: "ctx2"},
					},
				}},
			},
		},
	}

	expanded := map[int]bool{0: true}
	_, positions := RenderContent(d, expanded, 0, styles, nil, nil)

	if len(positions) != 1 {
		t.Fatalf("expected 1 file position, got %d", len(positions))
	}

	if !positions[0].Expanded {
		t.Error("file should be expanded")
	}

	// StartLine should be 0
	if positions[0].StartLine != 0 {
		t.Errorf("expected start line 0, got %d", positions[0].StartLine)
	}

	// EndLine should account for: header(1) + hunk_header(1) + 3 lines + trailing newline(1) = 6
	expectedEnd := 6
	if positions[0].EndLine != expectedEnd {
		t.Errorf("expected end line %d, got %d", expectedEnd, positions[0].EndLine)
	}
}

func TestRenderContentBinaryFileSkipsHunks(t *testing.T) {
	styles := ContentStyles{}
	d := &Diff{
		Files: []FileDiff{
			{
				OldName:  "img.png",
				NewName:  "img.png",
				IsBinary: true,
				Hunks: []Hunk{{
					Header: "@@ -1 +1 @@",
					Lines: []Line{
						{Type: DiffAdd, Content: "should be skipped"},
					},
				}},
			},
		},
	}

	// Even when expanded, binary files should not render hunk content
	expanded := map[int]bool{0: true}
	_, positions := RenderContent(d, expanded, 0, styles, nil, nil)

	// Binary file expanded should only show header, no hunk content
	// StartLine = 0, EndLine = 1 (just the header)
	if positions[0].EndLine != 1 {
		t.Errorf("binary file expanded should only show header (end=1), got end=%d", positions[0].EndLine)
	}
}

func TestRenderFileHeaderFilename(t *testing.T) {
	tests := []struct {
		name     string
		file     FileDiff
		wantName string
	}{
		{
			name:     "normal file",
			file:     FileDiff{OldName: "old.go", NewName: "new.go"},
			wantName: "new.go",
		},
		{
			name:     "deleted file uses old name",
			file:     FileDiff{OldName: "deleted.go", NewName: "/dev/null", IsDelete: true},
			wantName: "deleted.go",
		},
		{
			name:     "empty new name uses old name",
			file:     FileDiff{OldName: "only_old.go", NewName: ""},
			wantName: "only_old.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styles := ContentStyles{}
			header := RenderFileHeader(tt.file, false, false, styles)
			// The header should contain the expected filename
			if !contains(header, tt.wantName) {
				t.Errorf("header %q does not contain expected filename %q", header, tt.wantName)
			}
		})
	}
}

func TestRenderFileHeaderIndicators(t *testing.T) {
	styles := ContentStyles{}

	// Collapsed indicator
	collapsed := RenderFileHeader(FileDiff{NewName: "f.go"}, false, false, styles)
	if !contains(collapsed, "▶") {
		t.Error("collapsed header should contain ▶")
	}

	// Expanded indicator
	expanded := RenderFileHeader(FileDiff{NewName: "f.go"}, false, true, styles)
	if !contains(expanded, "▼") {
		t.Error("expanded header should contain ▼")
	}
}

func TestRenderFileHeaderBinaryStats(t *testing.T) {
	styles := ContentStyles{}
	header := RenderFileHeader(FileDiff{NewName: "img.png", IsBinary: true}, false, false, styles)
	if !contains(header, "(binary)") {
		t.Errorf("binary file header should contain '(binary)', got %q", header)
	}
}

// contains checks if s contains substr, stripping ANSI codes
func contains(s, substr string) bool {
	// Simple check - lipgloss may add ANSI codes in rendered output
	// but with zero-value styles, output should be plain
	return len(s) > 0 && len(substr) > 0 && (s == substr || containsPlain(s, substr))
}

func containsPlain(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
