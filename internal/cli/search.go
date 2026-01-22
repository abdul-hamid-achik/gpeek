package cli

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/agent"
	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/spf13/cobra"
)

// SearchResponse represents the JSON response for search command
type SearchResponse struct {
	Query   string             `json:"query"`
	Results []SearchResultInfo `json:"results"`
	Total   int                `json:"total"`
}

type SearchResultInfo struct {
	Hash       string  `json:"hash"`
	Message    string  `json:"message"`
	Author     string  `json:"author"`
	Similarity float64 `json:"similarity"`
	MatchedOn  string  `json:"matched_on"`
}

// IndexResponse represents the JSON response for index command
type IndexResponse struct {
	Indexed int    `json:"indexed"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

// SimilarResponse represents the JSON response for similar command
type SimilarResponse struct {
	Source  string             `json:"source"`
	Results []SearchResultInfo `json:"results"`
	Total   int                `json:"total"`
}

var (
	searchLimit int
	indexLimit  int
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Semantic search",
	Long:  `Perform semantic search on indexed content.`,
}

var searchCommitsCmd = &cobra.Command{
	Use:   "commits <query>",
	Short: "Search commits semantically",
	Long:  `Search commits using semantic similarity. Requires commits to be indexed first with 'gpeek index commits'.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSearchCommits,
}

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index content for semantic search",
	Long:  `Index repository content for semantic search capabilities.`,
}

var indexCommitsCmd = &cobra.Command{
	Use:   "commits",
	Short: "Index commits for semantic search",
	Long:  `Index commit history for semantic search. Creates embeddings from commit messages and changed files.`,
	RunE:  runIndexCommits,
}

var similarCmd = &cobra.Command{
	Use:   "similar",
	Short: "Find similar changes",
	Long:  `Find commits with similar changes to current staged changes or a specific file.`,
	RunE:  runSimilar,
}

var (
	similarStaged bool
	similarFile   string
)

func init() {
	searchCommitsCmd.Flags().IntVarP(&searchLimit, "limit", "n", 10, "Maximum number of results")
	searchCmd.AddCommand(searchCommitsCmd)

	indexCommitsCmd.Flags().IntVarP(&indexLimit, "limit", "n", 500, "Maximum number of commits to index")
	indexCmd.AddCommand(indexCommitsCmd)

	similarCmd.Flags().BoolVar(&similarStaged, "staged", false, "Find commits similar to staged changes")
	similarCmd.Flags().StringVar(&similarFile, "file", "", "Find commits similar to a specific file")
	similarCmd.Flags().IntVarP(&searchLimit, "limit", "n", 10, "Maximum number of results")

	// Add to root command
	RootCmd.AddCommand(searchCmd)
	RootCmd.AddCommand(indexCmd)
	RootCmd.AddCommand(similarCmd)
}

func runSearchCommits(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	vecgrep := agent.NewVecgrepIntegration(repo.Path())
	results, err := vecgrep.SearchCommits(query, searchLimit)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Convert to response type
	resultInfos := make([]SearchResultInfo, len(results))
	for i, r := range results {
		resultInfos[i] = SearchResultInfo{
			Hash:       r.Hash,
			Message:    r.Message,
			Author:     r.Author,
			Similarity: r.Similarity,
			MatchedOn:  r.MatchedOn,
		}
	}

	response := SearchResponse{
		Query:   query,
		Results: resultInfos,
		Total:   len(resultInfos),
	}

	output(response, formatSearchPlain)
	return nil
}

func runIndexCommits(cmd *cobra.Command, args []string) error {
	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	vecgrep := agent.NewVecgrepIntegration(repo.Path())
	if err := vecgrep.IndexCommits(repo, indexLimit); err != nil {
		return fmt.Errorf("indexing failed: %w", err)
	}

	response := IndexResponse{
		Indexed: indexLimit,
		Path:    repo.Path(),
		Message: fmt.Sprintf("Indexed up to %d commits successfully", indexLimit),
	}

	output(response, formatIndexPlain)
	return nil
}

func runSimilar(cmd *cobra.Command, args []string) error {
	if !similarStaged && similarFile == "" {
		return fmt.Errorf("either --staged or --file must be specified")
	}

	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	vecgrep := agent.NewVecgrepIntegration(repo.Path())

	var results []agent.SearchResult
	var source string

	if similarStaged {
		source = "staged changes"
		results, err = vecgrep.FindSimilarToStaged(repo, searchLimit)
	} else {
		source = similarFile
		results, err = vecgrep.FindSimilarToFile(repo, similarFile, searchLimit)
	}

	if err != nil {
		return fmt.Errorf("similar search failed: %w", err)
	}

	// Convert to response type
	resultInfos := make([]SearchResultInfo, len(results))
	for i, r := range results {
		resultInfos[i] = SearchResultInfo{
			Hash:       r.Hash,
			Message:    r.Message,
			Author:     r.Author,
			Similarity: r.Similarity,
			MatchedOn:  r.MatchedOn,
		}
	}

	response := SimilarResponse{
		Source:  source,
		Results: resultInfos,
		Total:   len(resultInfos),
	}

	output(response, formatSimilarPlain)
	return nil
}

func formatSearchPlain(data interface{}) string {
	response := data.(SearchResponse)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Search results for: %s\n\n", response.Query))

	if response.Total == 0 {
		sb.WriteString("No results found\n")
		return sb.String()
	}

	for _, r := range response.Results {
		sb.WriteString(fmt.Sprintf("%s (%.2f) - %s\n", r.Hash, r.Similarity, r.Message))
		if r.Author != "" {
			sb.WriteString(fmt.Sprintf("  by %s\n", r.Author))
		}
	}

	sb.WriteString(fmt.Sprintf("\nTotal: %d results\n", response.Total))
	return sb.String()
}

func formatIndexPlain(data interface{}) string {
	response := data.(IndexResponse)
	return fmt.Sprintf("%s\nPath: %s\n", response.Message, response.Path)
}

func formatSimilarPlain(data interface{}) string {
	response := data.(SimilarResponse)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Commits similar to: %s\n\n", response.Source))

	if response.Total == 0 {
		sb.WriteString("No similar commits found\n")
		return sb.String()
	}

	for _, r := range response.Results {
		sb.WriteString(fmt.Sprintf("%s (%.2f) - %s\n", r.Hash, r.Similarity, r.Message))
		if r.Author != "" {
			sb.WriteString(fmt.Sprintf("  by %s\n", r.Author))
		}
	}

	sb.WriteString(fmt.Sprintf("\nTotal: %d results\n", response.Total))
	return sb.String()
}
