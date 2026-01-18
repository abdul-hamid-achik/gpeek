package diff

import (
	"regexp"
	"strconv"
	"strings"
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

type Line struct {
	Type      DiffType
	Content   string
	OldNumber int
	NewNumber int
}

type Hunk struct {
	OldStart  int
	OldCount  int
	NewStart  int
	NewCount  int
	Header    string
	Lines     []Line
}

type FileDiff struct {
	OldName  string
	NewName  string
	IsBinary bool
	IsNew    bool
	IsDelete bool
	IsRename bool
	Hunks    []Hunk
}

type Diff struct {
	Files []FileDiff
}

var (
	diffHeaderRegex = regexp.MustCompile(`^diff --git a/(.*) b/(.*)$`)
	hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@(.*)$`)
	oldFileRegex    = regexp.MustCompile(`^--- (?:a/)?(.*)$`)
	newFileRegex    = regexp.MustCompile(`^\+\+\+ (?:b/)?(.*)$`)
)

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
		if strings.HasPrefix(line, "+") {
			diffLine = Line{
				Type:      DiffAdd,
				Content:   strings.TrimPrefix(line, "+"),
				NewNumber: newLineNum,
			}
			newLineNum++
		} else if strings.HasPrefix(line, "-") {
			diffLine = Line{
				Type:      DiffRemove,
				Content:   strings.TrimPrefix(line, "-"),
				OldNumber: oldLineNum,
			}
			oldLineNum++
		} else if strings.HasPrefix(line, " ") {
			diffLine = Line{
				Type:      DiffContext,
				Content:   strings.TrimPrefix(line, " "),
				OldNumber: oldLineNum,
				NewNumber: newLineNum,
			}
			oldLineNum++
			newLineNum++
		} else if line == "" {
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
