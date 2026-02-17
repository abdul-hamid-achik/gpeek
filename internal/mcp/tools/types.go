package tools

import (
	"fmt"
	"time"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/git"
)

// RepositoryInfo contains basic repository information
type RepositoryInfo struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Branch string `json:"branch"`
}

// FileInfo represents a file with its status
type FileInfo struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

// CommitInfo represents commit information
type CommitInfo struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"short_hash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Time      time.Time `json:"time"`
	TimeAgo   string    `json:"time_ago"`
	IsMerge   bool      `json:"is_merge"`
	Parents   []string  `json:"parents,omitempty"`
}

// DiffResponse is the response for diff queries
type DiffResponse struct {
	File   string         `json:"file,omitempty"`
	Commit string         `json:"commit,omitempty"`
	Staged bool           `json:"staged"`
	Files  []FileDiffInfo `json:"files"`
	Stats  DiffStats      `json:"stats"`
}

// FileDiffInfo represents a file diff
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

// HunkInfo represents a diff hunk
type HunkInfo struct {
	OldStart int        `json:"old_start"`
	OldCount int        `json:"old_count"`
	NewStart int        `json:"new_start"`
	NewCount int        `json:"new_count"`
	Header   string     `json:"header"`
	Lines    []LineInfo `json:"lines"`
}

// LineInfo represents a line in a diff
type LineInfo struct {
	Type      string     `json:"type"`
	Content   string     `json:"content"`
	OldNumber int        `json:"old_number,omitempty"`
	NewNumber int        `json:"new_number,omitempty"`
	Blame     *BlameInfo `json:"blame,omitempty"` // Optional blame info
}

// BlameInfo contains blame info for a line
type BlameInfo struct {
	Author  string `json:"author"`
	Hash    string `json:"hash"`
	TimeAgo string `json:"time_ago"`
}

// DiffStats contains diff statistics
type DiffStats struct {
	FilesChanged int `json:"files_changed"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
}

// LogResponse is the response for log queries
type LogResponse struct {
	Commits []CommitInfo `json:"commits"`
	Total   int          `json:"total"`
}

// SummaryResponse is the complete repository summary
type SummaryResponse struct {
	Repository    RepositoryInfo   `json:"repository"`
	Status        SummaryStatus    `json:"status"`
	RecentCommits []CommitInfo     `json:"recent_commits"`
	Branches      SummaryBranches  `json:"branches"`
	Stashes       SummaryStashes   `json:"stashes"`
	Tags          SummaryTags      `json:"tags"`
	Enhanced      *EnhancedSummary `json:"enhanced,omitempty"`
}

// EnhancedSummary contains additional analysis data
type EnhancedSummary struct {
	HotFiles    []HotFile       `json:"hot_files"`
	Languages   []LanguageInfo  `json:"languages"`
	ProjectType string          `json:"project_type"`
	Suggestions []string        `json:"suggestions,omitempty"`
}

// HotFile represents a frequently changed file
type HotFile struct {
	Path        string   `json:"path"`
	ChangeCount int      `json:"change_count"`
	Authors     []string `json:"authors"`
}

// LanguageInfo represents detected language statistics
type LanguageInfo struct {
	Name       string `json:"name"`
	FileCount  int    `json:"file_count"`
	Percentage float64 `json:"percentage"`
}

// SummaryStatus is the status section of summary
type SummaryStatus struct {
	Staged         []FileInfo `json:"staged"`
	Unstaged       []FileInfo `json:"unstaged"`
	Untracked      []string   `json:"untracked"`
	StagedCount    int        `json:"staged_count"`
	UnstagedCount  int        `json:"unstaged_count"`
	UntrackedCount int        `json:"untracked_count"`
	IsClean        bool       `json:"is_clean"`
	HasConflicts   bool       `json:"has_conflicts"`
}

// SummaryBranches is the branches section of summary
type SummaryBranches struct {
	Current string       `json:"current"`
	Local   []BranchInfo `json:"local"`
	Count   int          `json:"count"`
}

// BranchInfo represents branch information
type BranchInfo struct {
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	IsCurrent bool   `json:"is_current"`
	Upstream  string `json:"upstream,omitempty"`
}

// SummaryStashes is the stashes section of summary
type SummaryStashes struct {
	Count   int         `json:"count"`
	Entries []StashInfo `json:"entries"`
}

// StashInfo represents stash information
type StashInfo struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
	Branch  string `json:"branch,omitempty"`
}

// SummaryTags is the tags section of summary
type SummaryTags struct {
	Count int       `json:"count"`
	Tags  []TagInfo `json:"tags"`
}

// TagInfo represents tag information
type TagInfo struct {
	Name        string `json:"name"`
	Hash        string `json:"hash"`
	IsAnnotated bool   `json:"is_annotated"`
}

// BlameResponse is the response for blame queries
type BlameResponse struct {
	File  string          `json:"file"`
	Lines []BlameLineInfo `json:"lines"`
	Total int             `json:"total"`
}

// BlameLineInfo represents a blame line
type BlameLineInfo struct {
	LineNum int       `json:"line_num"`
	Hash    string    `json:"hash,omitempty"`
	Author  string    `json:"author,omitempty"`
	Time    time.Time `json:"time,omitempty"`
	TimeAgo string    `json:"time_ago,omitempty"`
	Content string    `json:"content"`
}

// BranchesResponse is the response for branches queries
type BranchesResponse struct {
	Current  string             `json:"current"`
	Branches []BranchDetailInfo `json:"branches"`
	Total    int                `json:"total"`
}

// BranchDetailInfo represents detailed branch information
type BranchDetailInfo struct {
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	ShortHash string `json:"short_hash"`
	IsRemote  bool   `json:"is_remote"`
	IsCurrent bool   `json:"is_current"`
	Upstream  string `json:"upstream,omitempty"`
}

// StashesResponse is the response for stashes queries
type StashesResponse struct {
	Stashes []StashDetailInfo `json:"stashes"`
	Total   int               `json:"total"`
}

// StashDetailInfo represents detailed stash information
type StashDetailInfo struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
	Branch  string `json:"branch,omitempty"`
	Hash    string `json:"hash"`
	TimeAgo string `json:"time_ago"`
}

// TagsResponse is the response for tags queries
type TagsResponse struct {
	Tags  []TagDetailInfo `json:"tags"`
	Total int             `json:"total"`
}

// TagDetailInfo represents detailed tag information
type TagDetailInfo struct {
	Name        string `json:"name"`
	Hash        string `json:"hash"`
	ShortHash   string `json:"short_hash"`
	Message     string `json:"message,omitempty"`
	Tagger      string `json:"tagger,omitempty"`
	TimeAgo     string `json:"time_ago,omitempty"`
	IsAnnotated bool   `json:"is_annotated"`
}

// StageResponse is the response for stage operations
type StageResponse struct {
	Staged []FileInfo `json:"staged"`
	Total  int        `json:"total"`
}

// UnstageResponse is the response for unstage operations
type UnstageResponse struct {
	Remaining []FileInfo `json:"remaining_staged"`
	Total     int        `json:"total"`
}

// CommitWriteResponse is the response for commit operations
type CommitWriteResponse struct {
	Hash      string `json:"hash"`
	ShortHash string `json:"short_hash"`
	Message   string `json:"message"`
}

// StashSaveResponse is the response for stash save operations
type StashSaveResponse struct {
	Reference string `json:"reference"`
	Message   string `json:"message"`
}

// StashOpResponse is the response for stash pop/drop operations
type StashOpResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Reference string `json:"reference"`
}

// DiscardResponse is the response for discard operations
type DiscardResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

// Builder functions

// BuildDiffResponse builds a DiffResponse from parsed diff
func BuildDiffResponse(parsed *diff.Diff, file, commit string, staged bool) DiffResponse {
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

	return DiffResponse{
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
}

// BuildLogResponse builds a LogResponse from commits
func BuildLogResponse(commits []git.Commit) LogResponse {
	infos := make([]CommitInfo, len(commits))
	for i, c := range commits {
		shortHash := c.Hash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}
		infos[i] = CommitInfo{
			Hash:      c.Hash,
			ShortHash: shortHash,
			Message:   c.Message,
			Author:    c.Author,
			Email:     c.Email,
			Time:      c.Time,
			TimeAgo:   TimeAgo(c.Time),
			IsMerge:   c.IsMerge,
			Parents:   c.Parents,
		}
	}
	return LogResponse{
		Commits: infos,
		Total:   len(infos),
	}
}

// BuildSummaryResponse builds a complete SummaryResponse
func BuildSummaryResponse(repo *git.Repository, commitLimit int) SummaryResponse {
	response := SummaryResponse{
		Repository: RepositoryInfo{
			Name:   repo.Name(),
			Path:   repo.Path(),
			Branch: repo.CurrentBranch(),
		},
	}

	// Get status
	status, err := repo.Status()
	if err == nil {
		staged := make([]FileInfo, len(status.Staged))
		for i, f := range status.Staged {
			staged[i] = FileInfo{Path: f.Path, Status: f.Status.String()}
		}
		unstaged := make([]FileInfo, len(status.Unstaged))
		for i, f := range status.Unstaged {
			unstaged[i] = FileInfo{Path: f.Path, Status: f.Status.String()}
		}

		hasConflicts := false
		for _, f := range status.Staged {
			if f.Status == git.StatusUpdatedButUnmerged {
				hasConflicts = true
				break
			}
		}

		response.Status = SummaryStatus{
			Staged:         staged,
			Unstaged:       unstaged,
			Untracked:      status.Untracked,
			StagedCount:    len(status.Staged),
			UnstagedCount:  len(status.Unstaged),
			UntrackedCount: len(status.Untracked),
			IsClean:        len(status.Staged) == 0 && len(status.Unstaged) == 0 && len(status.Untracked) == 0,
			HasConflicts:   hasConflicts,
		}
	}

	// Get recent commits
	commits, err := repo.Log(commitLimit)
	if err == nil {
		commitInfos := make([]CommitInfo, len(commits))
		for i, c := range commits {
			shortHash := c.Hash
			if len(shortHash) > 7 {
				shortHash = shortHash[:7]
			}
			commitInfos[i] = CommitInfo{
				Hash:      c.Hash,
				ShortHash: shortHash,
				Message:   c.Message,
				Author:    c.Author,
				Email:     c.Email,
				Time:      c.Time,
				TimeAgo:   TimeAgo(c.Time),
				IsMerge:   c.IsMerge,
			}
		}
		response.RecentCommits = commitInfos
	}

	// Get branches
	branches, err := repo.ListBranches()
	if err == nil {
		branchInfos := make([]BranchInfo, len(branches))
		for i, b := range branches {
			branchInfos[i] = BranchInfo{
				Name:      b.Name,
				Hash:      b.Hash,
				IsCurrent: b.IsCurrent,
				Upstream:  b.Upstream,
			}
		}
		response.Branches = SummaryBranches{
			Current: repo.CurrentBranch(),
			Local:   branchInfos,
			Count:   len(branches),
		}
	}

	// Get stashes
	stashes, err := repo.StashList()
	if err == nil {
		stashInfos := make([]StashInfo, len(stashes))
		for i, s := range stashes {
			stashInfos[i] = StashInfo{
				Index:   s.Index,
				Message: s.Message,
				Branch:  s.Branch,
			}
		}
		response.Stashes = SummaryStashes{
			Count:   len(stashes),
			Entries: stashInfos,
		}
	}

	// Get tags
	tags, err := repo.ListTags()
	if err == nil {
		tagInfos := make([]TagInfo, len(tags))
		for i, t := range tags {
			tagInfos[i] = TagInfo{
				Name:        t.Name,
				Hash:        t.Hash,
				IsAnnotated: t.IsAnnotated,
			}
		}
		response.Tags = SummaryTags{
			Count: len(tags),
			Tags:  tagInfos,
		}
	}

	return response
}

// BuildBlameResponse builds a BlameResponse
func BuildBlameResponse(file string, lines []git.BlameLine) BlameResponse {
	infos := make([]BlameLineInfo, len(lines))
	for i, l := range lines {
		info := BlameLineInfo{
			LineNum: l.LineNum,
			Hash:    l.Hash,
			Author:  l.Author,
			Time:    l.Time,
			Content: l.Content,
		}
		if !l.Time.IsZero() {
			info.TimeAgo = TimeAgo(l.Time)
		}
		if len(info.Hash) > 8 {
			info.Hash = info.Hash[:8]
		}
		infos[i] = info
	}
	return BlameResponse{
		File:  file,
		Lines: infos,
		Total: len(infos),
	}
}

// BuildBranchesResponse builds a BranchesResponse
func BuildBranchesResponse(current string, branches []git.Branch) BranchesResponse {
	infos := make([]BranchDetailInfo, len(branches))
	for i, b := range branches {
		shortHash := b.Hash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}
		infos[i] = BranchDetailInfo{
			Name:      b.Name,
			Hash:      b.Hash,
			ShortHash: shortHash,
			IsRemote:  b.IsRemote,
			IsCurrent: b.IsCurrent,
			Upstream:  b.Upstream,
		}
	}
	return BranchesResponse{
		Current:  current,
		Branches: infos,
		Total:    len(infos),
	}
}

// BuildStashesResponse builds a StashesResponse
func BuildStashesResponse(stashes []git.Stash) StashesResponse {
	infos := make([]StashDetailInfo, len(stashes))
	for i, s := range stashes {
		infos[i] = StashDetailInfo{
			Index:   s.Index,
			Message: s.Message,
			Branch:  s.Branch,
			Hash:    s.Hash,
			TimeAgo: TimeAgo(s.Time),
		}
	}
	return StashesResponse{
		Stashes: infos,
		Total:   len(infos),
	}
}

// BuildTagsResponse builds a TagsResponse
func BuildTagsResponse(tags []git.Tag) TagsResponse {
	infos := make([]TagDetailInfo, len(tags))
	for i, t := range tags {
		shortHash := t.Hash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}
		info := TagDetailInfo{
			Name:        t.Name,
			Hash:        t.Hash,
			ShortHash:   shortHash,
			Message:     t.Message,
			Tagger:      t.Tagger,
			IsAnnotated: t.IsAnnotated,
		}
		if !t.TaggerTime.IsZero() {
			info.TimeAgo = TimeAgo(t.TaggerTime)
		}
		infos[i] = info
	}
	return TagsResponse{
		Tags:  infos,
		Total: len(infos),
	}
}

// TimeAgo returns a human-readable time difference
func TimeAgo(t time.Time) string {
	duration := time.Since(t)
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 30*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case duration < 365*24*time.Hour:
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
