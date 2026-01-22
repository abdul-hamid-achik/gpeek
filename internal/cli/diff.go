package cli

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/spf13/cobra"
)

// DiffResponse represents the JSON response for diff command
type DiffResponse struct {
	File    string         `json:"file,omitempty"`
	Commit  string         `json:"commit,omitempty"`
	Staged  bool           `json:"staged"`
	Files   []FileDiffInfo `json:"files"`
	Stats   DiffStats      `json:"stats"`
	RawDiff string         `json:"raw_diff,omitempty"`
}

type FileDiffInfo struct {
	OldName   string     `json:"old_name"`
	NewName   string     `json:"new_name"`
	IsBinary  bool       `json:"is_binary"`
	IsNew     bool       `json:"is_new"`
	IsDelete  bool       `json:"is_delete"`
	IsRename  bool       `json:"is_rename"`
	Hunks     []HunkInfo `json:"hunks"`
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
}

type HunkInfo struct {
	OldStart int        `json:"old_start"`
	OldCount int        `json:"old_count"`
	NewStart int        `json:"new_start"`
	NewCount int        `json:"new_count"`
	Header   string     `json:"header"`
	Lines    []LineInfo `json:"lines"`
}

type LineInfo struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	OldNumber int    `json:"old_number,omitempty"`
	NewNumber int    `json:"new_number,omitempty"`
}

type DiffStats struct {
	FilesChanged int `json:"files_changed"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
}

var (
	diffStaged     bool
	diffCommit     string
	diffIncludeRaw bool
)

var diffCmd = &cobra.Command{
	Use:   "diff [file]",
	Short: "Show changes in working tree or commit",
	Long:  `Display structured diff for files, staged changes, or commits.`,
	RunE:  runDiff,
}

func init() {
	diffCmd.Flags().BoolVarP(&diffStaged, "staged", "s", false, "Show staged changes")
	diffCmd.Flags().StringVarP(&diffCommit, "commit", "c", "", "Show changes for a specific commit")
	diffCmd.Flags().BoolVar(&diffIncludeRaw, "raw", false, "Include raw diff output")
}

func runDiff(cmd *cobra.Command, args []string) error {
	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	var rawDiff string
	var file string

	if len(args) > 0 {
		file = args[0]
	}

	if diffCommit != "" {
		// Get commit diff
		rawDiff, err = repo.CommitDiff(diffCommit)
		if err != nil {
			return fmt.Errorf("failed to get commit diff: %w", err)
		}
	} else if file != "" {
		// Get file diff
		rawDiff, err = repo.FileDiff(file, diffStaged)
		if err != nil {
			return fmt.Errorf("failed to get file diff: %w", err)
		}
	} else {
		// Get all diffs (staged or unstaged)
		status, err := repo.Status()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		var files []git.FileEntry
		if diffStaged {
			files = status.Staged
		} else {
			files = status.Unstaged
		}

		var allDiffs strings.Builder
		for _, f := range files {
			d, err := repo.FileDiff(f.Path, diffStaged)
			if err != nil {
				continue
			}
			allDiffs.WriteString(d)
		}
		rawDiff = allDiffs.String()
	}

	// Parse the diff
	parsed := diff.Parse(rawDiff)

	response := buildDiffResponse(parsed, file, diffCommit, diffStaged, rawDiff)
	output(response, formatDiffPlain)
	return nil
}

func buildDiffResponse(parsed *diff.Diff, file, commit string, staged bool, rawDiff string) DiffResponse {
	files := make([]FileDiffInfo, len(parsed.Files))
	totalAdded, totalRemoved := 0, 0

	for i, f := range parsed.Files {
		hunks := make([]HunkInfo, len(f.Hunks))
		fileAdded, fileRemoved := 0, 0

		for j, h := range f.Hunks {
			lines := make([]LineInfo, len(h.Lines))
			for k, l := range h.Lines {
				lines[k] = LineInfo{
					Type:      l.Type.String(),
					Content:   l.Content,
					OldNumber: l.OldNumber,
					NewNumber: l.NewNumber,
				}
				switch l.Type {
				case diff.DiffAdd:
					fileAdded++
				case diff.DiffRemove:
					fileRemoved++
				}
			}
			hunks[j] = HunkInfo{
				OldStart: h.OldStart,
				OldCount: h.OldCount,
				NewStart: h.NewStart,
				NewCount: h.NewCount,
				Header:   h.Header,
				Lines:    lines,
			}
		}

		files[i] = FileDiffInfo{
			OldName:   f.OldName,
			NewName:   f.NewName,
			IsBinary:  f.IsBinary,
			IsNew:     f.IsNew,
			IsDelete:  f.IsDelete,
			IsRename:  f.IsRename,
			Hunks:     hunks,
			Additions: fileAdded,
			Deletions: fileRemoved,
		}
		totalAdded += fileAdded
		totalRemoved += fileRemoved
	}

	response := DiffResponse{
		File:   file,
		Commit: commit,
		Staged: staged,
		Files:  files,
		Stats: DiffStats{
			FilesChanged: len(files),
			Additions:    totalAdded,
			Deletions:    totalRemoved,
		},
	}

	if diffIncludeRaw {
		response.RawDiff = rawDiff
	}

	return response
}

func formatDiffPlain(data interface{}) string {
	response := data.(DiffResponse)
	var sb strings.Builder

	if response.Commit != "" {
		sb.WriteString(fmt.Sprintf("Diff for commit %s\n", response.Commit))
	} else if response.File != "" {
		sb.WriteString(fmt.Sprintf("Diff for %s", response.File))
		if response.Staged {
			sb.WriteString(" (staged)")
		}
		sb.WriteString("\n")
	} else {
		if response.Staged {
			sb.WriteString("Staged changes:\n")
		} else {
			sb.WriteString("Unstaged changes:\n")
		}
	}

	for _, f := range response.Files {
		name := f.NewName
		if f.IsDelete {
			name = f.OldName
		}
		sb.WriteString(fmt.Sprintf("\n--- %s ---\n", name))

		if f.IsBinary {
			sb.WriteString("Binary file\n")
			continue
		}

		for _, h := range f.Hunks {
			sb.WriteString(fmt.Sprintf("%s\n", h.Header))
			for _, l := range h.Lines {
				prefix := " "
				switch l.Type {
				case "add":
					prefix = "+"
				case "remove":
					prefix = "-"
				}
				sb.WriteString(fmt.Sprintf("%s%s\n", prefix, l.Content))
			}
		}
	}

	sb.WriteString(fmt.Sprintf("\n%d files changed, %d insertions(+), %d deletions(-)\n",
		response.Stats.FilesChanged, response.Stats.Additions, response.Stats.Deletions))

	return sb.String()
}
