package diff

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type DiffType int

const (
	DiffContext DiffType = iota
	DiffAdd
	DiffRemove
	DiffMeta
	DiffHunk
	DiffBinary
)

// String returns the string representation of DiffType
func (t DiffType) String() string {
	switch t {
	case DiffContext:
		return "context"
	case DiffAdd:
		return "add"
	case DiffRemove:
		return "remove"
	case DiffMeta:
		return "meta"
	case DiffHunk:
		return "hunk"
	case DiffBinary:
		return "binary"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler for DiffType
func (t DiffType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

type Line struct {
	Type      DiffType `json:"type"`
	Content   string   `json:"content"`
	OldNumber int      `json:"old_number,omitempty"`
	NewNumber int      `json:"new_number,omitempty"`
}

type Hunk struct {
	OldStart int    `json:"old_start"`
	OldCount int    `json:"old_count"`
	NewStart int    `json:"new_start"`
	NewCount int    `json:"new_count"`
	Header   string `json:"header"`
	Lines    []Line `json:"lines"`
}

type FileDiff struct {
	OldName  string `json:"old_name"`
	NewName  string `json:"new_name"`
	IsBinary bool   `json:"is_binary"`
	IsNew    bool   `json:"is_new"`
	IsDelete bool   `json:"is_delete"`
	IsRename bool   `json:"is_rename"`
	Hunks    []Hunk `json:"hunks"`
}

type Diff struct {
	Files []FileDiff `json:"files"`
}

var (
	diffHeaderRegex = regexp.MustCompile(`^diff --git a/(.*) b/(.*)$`)
	hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@(.*)$`)
	oldFileRegex    = regexp.MustCompile(`^--- (?:a/)?(.*)$`)
	newFileRegex    = regexp.MustCompile(`^\+\+\+ (?:b/)?(.*)$`)
)

// isBinaryContent checks if content appears to be binary
func isBinaryContent(content string) bool {
	// Check for null bytes (strong indicator of binary)
	if strings.Contains(content, "\x00") {
		return true
	}
	// Check if content is valid UTF-8
	if !utf8.ValidString(content) {
		return true
	}
	return false
}

func Parse(input string) *Diff {
	diff := &Diff{}
	lines := strings.Split(input, "\n")

	var currentFile *FileDiff
	var currentHunk *Hunk
	var oldLineNum, newLineNum int

	for _, line := range lines {
		if matches := diffHeaderRegex.FindStringSubmatch(line); matches != nil {
			if currentFile != nil {
				if currentHunk != nil {
					currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
				}
				diff.Files = append(diff.Files, *currentFile)
			}
			currentFile = &FileDiff{
				OldName: matches[1],
				NewName: matches[2],
			}
			currentHunk = nil
			continue
		}

		if currentFile == nil {
			continue
		}

		if matches := oldFileRegex.FindStringSubmatch(line); matches != nil {
			if matches[1] == "/dev/null" {
				currentFile.IsNew = true
			}
			continue
		}

		if matches := newFileRegex.FindStringSubmatch(line); matches != nil {
			if matches[1] == "/dev/null" {
				currentFile.IsDelete = true
			}
			continue
		}

		if strings.HasPrefix(line, "Binary files") {
			currentFile.IsBinary = true
			continue
		}

		if strings.HasPrefix(line, "rename from") || strings.HasPrefix(line, "rename to") {
			currentFile.IsRename = true
			continue
		}

		if matches := hunkHeaderRegex.FindStringSubmatch(line); matches != nil {
			if currentHunk != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}

			oldStart, _ := strconv.Atoi(matches[1])
			oldCount := 1
			if matches[2] != "" {
				oldCount, _ = strconv.Atoi(matches[2])
			}
			newStart, _ := strconv.Atoi(matches[3])
			newCount := 1
			if matches[4] != "" {
				newCount, _ = strconv.Atoi(matches[4])
			}

			currentHunk = &Hunk{
				OldStart: oldStart,
				OldCount: oldCount,
				NewStart: newStart,
				NewCount: newCount,
				Header:   line,
			}
			oldLineNum = oldStart
			newLineNum = newStart
			continue
		}

		if currentHunk == nil {
			continue
		}

		var diffLine Line
		var lineContent string
		if strings.HasPrefix(line, "+") {
			lineContent = strings.TrimPrefix(line, "+")
			diffLine = Line{
				Type:      DiffAdd,
				Content:   lineContent,
				NewNumber: newLineNum,
			}
			newLineNum++
		} else if strings.HasPrefix(line, "-") {
			lineContent = strings.TrimPrefix(line, "-")
			diffLine = Line{
				Type:      DiffRemove,
				Content:   lineContent,
				OldNumber: oldLineNum,
			}
			oldLineNum++
		} else if strings.HasPrefix(line, " ") {
			lineContent = strings.TrimPrefix(line, " ")
			diffLine = Line{
				Type:      DiffContext,
				Content:   lineContent,
				OldNumber: oldLineNum,
				NewNumber: newLineNum,
			}
			oldLineNum++
			newLineNum++
		} else if line == "" {
			lineContent = ""
			diffLine = Line{
				Type:      DiffContext,
				Content:   "",
				OldNumber: oldLineNum,
				NewNumber: newLineNum,
			}
			oldLineNum++
			newLineNum++
		} else {
			continue
		}

		// Check for binary content and mark file accordingly
		if lineContent != "" && isBinaryContent(lineContent) {
			currentFile.IsBinary = true
		}

		currentHunk.Lines = append(currentHunk.Lines, diffLine)
	}

	if currentFile != nil {
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
		}
		diff.Files = append(diff.Files, *currentFile)
	}

	return diff
}

func (d *Diff) Stats() (added, removed int) {
	for _, file := range d.Files {
		for _, hunk := range file.Hunks {
			for _, line := range hunk.Lines {
				switch line.Type {
				case DiffAdd:
					added++
				case DiffRemove:
					removed++
				}
			}
		}
	}
	return
}

func (d *Diff) FileCount() int {
	return len(d.Files)
}
