package git

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// HotFile represents a frequently changed file
type HotFile struct {
	Path        string
	ChangeCount int
	Authors     []string
}

// LanguageStats represents language detection results
type LanguageStats struct {
	Name       string
	FileCount  int
	Percentage float64
}

// RepoAnalysis contains repository analysis results
type RepoAnalysis struct {
	HotFiles    []HotFile
	Languages   []LanguageStats
	ProjectType string
}

// AnalyzeRepository performs enhanced analysis of the repository
func (r *Repository) AnalyzeRepository(commitLimit int) (*RepoAnalysis, error) {
	analysis := &RepoAnalysis{}

	// Analyze hot files
	hotFiles, err := r.findHotFiles(commitLimit)
	if err == nil {
		analysis.HotFiles = hotFiles
	}

	// Detect languages
	languages, err := r.detectLanguages()
	if err == nil {
		analysis.Languages = languages
	}

	// Detect project type
	analysis.ProjectType = r.detectProjectType()

	return analysis, nil
}

// findHotFiles finds files with most changes in recent commits
func (r *Repository) findHotFiles(commitLimit int) ([]HotFile, error) {
	fileChanges := make(map[string]int)
	fileAuthors := make(map[string]map[string]bool)

	head, err := r.repo.Head()
	if err != nil {
		return nil, err
	}

	iter, err := r.repo.Log(&gogit.LogOptions{
		From:  head.Hash(),
		Order: gogit.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, err
	}

	count := 0
	_ = iter.ForEach(func(c *object.Commit) error {
		if count >= commitLimit {
			return nil
		}
		count++

		// Get files changed in this commit
		stats, err := c.Stats()
		if err != nil {
			return nil
		}

		for _, stat := range stats {
			fileChanges[stat.Name]++
			if fileAuthors[stat.Name] == nil {
				fileAuthors[stat.Name] = make(map[string]bool)
			}
			fileAuthors[stat.Name][c.Author.Name] = true
		}

		return nil
	})

	// Convert to slice and sort by change count
	type fileChange struct {
		path  string
		count int
	}
	var changes []fileChange
	for path, count := range fileChanges {
		changes = append(changes, fileChange{path, count})
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[j].count < changes[i].count
	})

	// Take top 10
	limit := 10
	if len(changes) < limit {
		limit = len(changes)
	}

	result := make([]HotFile, limit)
	for i := 0; i < limit; i++ {
		var authors []string
		for author := range fileAuthors[changes[i].path] {
			authors = append(authors, author)
		}
		result[i] = HotFile{
			Path:        changes[i].path,
			ChangeCount: changes[i].count,
			Authors:     authors,
		}
	}

	return result, nil
}

// detectLanguages detects programming languages in the repository
func (r *Repository) detectLanguages() ([]LanguageStats, error) {
	langMap := make(map[string]int)
	totalFiles := 0

	err := filepath.Walk(r.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden and common non-code directories
		if info.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" || base == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		// Map extension to language
		ext := strings.ToLower(filepath.Ext(path))
		lang := extensionToLanguage(ext)
		if lang != "" {
			langMap[lang]++
			totalFiles++
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert to slice
	var languages []LanguageStats
	for name, count := range langMap {
		pct := 0.0
		if totalFiles > 0 {
			pct = float64(count) / float64(totalFiles) * 100
		}
		languages = append(languages, LanguageStats{
			Name:       name,
			FileCount:  count,
			Percentage: pct,
		})
	}

	// Sort by file count
	sort.Slice(languages, func(i, j int) bool {
		return languages[j].FileCount < languages[i].FileCount
	})

	// Limit to top 10
	if len(languages) > 10 {
		languages = languages[:10]
	}

	return languages, nil
}

// extToLang maps file extensions to language names (initialized once at package level)
var extToLang = map[string]string{
	".go":     "Go",
	".py":     "Python",
	".js":     "JavaScript",
	".ts":     "TypeScript",
	".jsx":    "JavaScript",
	".tsx":    "TypeScript",
	".java":   "Java",
	".kt":     "Kotlin",
	".rb":     "Ruby",
	".rs":     "Rust",
	".c":      "C",
	".cpp":    "C++",
	".cc":     "C++",
	".h":      "C/C++ Header",
	".hpp":    "C++ Header",
	".cs":     "C#",
	".swift":  "Swift",
	".php":    "PHP",
	".scala":  "Scala",
	".sh":     "Shell",
	".bash":   "Shell",
	".zsh":    "Shell",
	".sql":    "SQL",
	".html":   "HTML",
	".css":    "CSS",
	".scss":   "SCSS",
	".sass":   "SASS",
	".less":   "LESS",
	".json":   "JSON",
	".yaml":   "YAML",
	".yml":    "YAML",
	".xml":    "XML",
	".md":     "Markdown",
	".vue":    "Vue",
	".svelte": "Svelte",
}

// extensionToLanguage maps file extensions to language names
func extensionToLanguage(ext string) string {
	return extToLang[ext]
}

// detectProjectType detects the type of project
func (r *Repository) detectProjectType() string {
	// Check for common project indicators
	indicators := map[string]string{
		"package.json":    "Node.js",
		"go.mod":          "Go",
		"Cargo.toml":      "Rust",
		"pom.xml":         "Maven/Java",
		"build.gradle":    "Gradle/Java",
		"requirements.txt": "Python",
		"pyproject.toml":  "Python",
		"Gemfile":         "Ruby",
		"composer.json":   "PHP",
		"Package.swift":   "Swift",
		"CMakeLists.txt":  "CMake/C++",
		"Makefile":        "Make",
	}

	for file, projectType := range indicators {
		if _, err := os.Stat(filepath.Join(r.path, file)); err == nil {
			return projectType
		}
	}

	return "Unknown"
}
