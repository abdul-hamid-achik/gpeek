package mcp

import (
	"fmt"
	"time"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/git"
)

// Response types for MCP tools

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

// StatusResponse is the response for status queries
type StatusResponse struct {
	Repository RepositoryInfo `json:"repository"`
	Staged     []FileInfo     `json:"staged"`
	Unstaged   []FileInfo     `json:"unstaged"`
	Untracked  []string       `json:"untracked"`
	Summary    StatusSummary  `json:"summary"`
}

// StatusSummary provides summary counts
type StatusSummary struct {
	StagedCount    int  `json:"staged_count"`
	UnstagedCount  int  `json:"unstaged_count"`
	UntrackedCount int  `json:"untracked_count"`
	IsClean        bool `json:"is_clean"`
	HasConflicts   bool `json:"has_conflicts"`
}

type DiffResponse struct {
	File   string         `json:"file,omitempty"`
	Commit string         `json:"commit,omitempty"`
	Staged bool           `json:"staged"`
	Files  []FileDiffInfo `json:"files"`
	Stats  DiffStats      `json:"stats"`
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

type LogResponse struct {
	Commits []CommitInfo `json:"commits"`
	Total   int          `json:"total"`
}

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

type SummaryResponse struct {
	Repository    RepositoryInfo  `json:"repository"`
	Status        SummaryStatus   `json:"status"`
	RecentCommits []CommitInfo    `json:"recent_commits"`
	Branches      SummaryBranches `json:"branches"`
	Stashes       SummaryStashes  `json:"stashes"`
	Tags          SummaryTags     `json:"tags"`
}

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

type SummaryBranches struct {
	Current string       `json:"current"`
	Local   []BranchInfo `json:"local"`
	Count   int          `json:"count"`
}

type BranchInfo struct {
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	IsCurrent bool   `json:"is_current"`
	Upstream  string `json:"upstream,omitempty"`
}

type SummaryStashes struct {
	Count   int         `json:"count"`
	Entries []StashInfo `json:"entries"`
}

type StashInfo struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
	Branch  string `json:"branch,omitempty"`
}

type SummaryTags struct {
	Count int       `json:"count"`
	Tags  []TagInfo `json:"tags"`
}

type TagInfo struct {
	Name        string `json:"name"`
	Hash        string `json:"hash"`
	IsAnnotated bool   `json:"is_annotated"`
}

type BlameResponse struct {
	File  string          `json:"file"`
	Lines []BlameLineInfo `json:"lines"`
	Total int             `json:"total"`
}

type BlameLineInfo struct {
	LineNum int       `json:"line_num"`
	Hash    string    `json:"hash,omitempty"`
	Author  string    `json:"author,omitempty"`
	Time    time.Time `json:"time,omitempty"`
	TimeAgo string    `json:"time_ago,omitempty"`
	Content string    `json:"content"`
}

type BranchesResponse struct {
	Current  string             `json:"current"`
	Branches []BranchDetailInfo `json:"branches"`
	Total    int                `json:"total"`
}

type BranchDetailInfo struct {
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	ShortHash string `json:"short_hash"`
	IsRemote  bool   `json:"is_remote"`
	IsCurrent bool   `json:"is_current"`
	Upstream  string `json:"upstream,omitempty"`
}

type StashesResponse struct {
	Stashes []StashDetailInfo `json:"stashes"`
	Total   int               `json:"total"`
}

type StashDetailInfo struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
	Branch  string `json:"branch,omitempty"`
	Hash    string `json:"hash"`
	TimeAgo string `json:"time_ago"`
}

type TagsResponse struct {
	Tags  []TagDetailInfo `json:"tags"`
	Total int             `json:"total"`
}

type TagDetailInfo struct {
	Name        string `json:"name"`
	Hash        string `json:"hash"`
	ShortHash   string `json:"short_hash"`
	Message     string `json:"message,omitempty"`
	Tagger      string `json:"tagger,omitempty"`
	TimeAgo     string `json:"time_ago,omitempty"`
	IsAnnotated bool   `json:"is_annotated"`
}

// Builder functions

func buildDiffResponse(parsed *diff.Diff, file, commit string, staged bool) DiffResponse {
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

func buildLogResponse(commits []git.Commit) LogResponse {
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
			TimeAgo:   timeAgo(c.Time),
			IsMerge:   c.IsMerge,
			Parents:   c.Parents,
		}
	}
	return LogResponse{
		Commits: infos,
		Total:   len(infos),
	}
}

func buildSummaryResponse(repo *git.Repository, commitLimit int) SummaryResponse {
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
				TimeAgo:   timeAgo(c.Time),
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

func buildBlameResponse(file string, lines []git.BlameLine) BlameResponse {
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
			info.TimeAgo = timeAgo(l.Time)
		}
		// Shorten hash for display
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

func buildBranchesResponse(current string, branches []git.Branch) BranchesResponse {
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

func buildStashesResponse(stashes []git.Stash) StashesResponse {
	infos := make([]StashDetailInfo, len(stashes))
	for i, s := range stashes {
		infos[i] = StashDetailInfo{
			Index:   s.Index,
			Message: s.Message,
			Branch:  s.Branch,
			Hash:    s.Hash,
			TimeAgo: timeAgo(s.Time),
		}
	}
	return StashesResponse{
		Stashes: infos,
		Total:   len(infos),
	}
}

func buildTagsResponse(tags []git.Tag) TagsResponse {
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
			info.TimeAgo = timeAgo(t.TaggerTime)
		}
		infos[i] = info
	}
	return TagsResponse{
		Tags:  infos,
		Total: len(infos),
	}
}

func timeAgo(t time.Time) string {
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
