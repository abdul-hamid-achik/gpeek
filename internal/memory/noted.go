package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ErrNotedUnavailable is returned when noted is not installed
var ErrNotedUnavailable = errors.New("noted is not installed - install from https://github.com/your-org/noted")

// NotedProvider provides memory operations via noted CLI
type NotedProvider struct {
	available bool
}

// NewNotedProvider creates a new noted provider
func NewNotedProvider() *NotedProvider {
	return &NotedProvider{
		available: checkNotedAvailable(),
	}
}

// checkNotedAvailable checks if noted is installed
func checkNotedAvailable() bool {
	_, err := exec.LookPath("noted")
	return err == nil
}

// Available returns true if noted is installed
func (n *NotedProvider) Available() bool {
	return n.available
}

// Memory represents a stored memory
type Memory struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// RememberOptions configures memory storage
type RememberOptions struct {
	Category string            // Category for the memory (e.g., "git", "code", "decision")
	Tags     []string          // Tags for retrieval
	Metadata map[string]string // Additional metadata
}

// Remember stores a memory using noted
func (n *NotedProvider) Remember(content string, opts RememberOptions) (*Memory, error) {
	if !n.available {
		return nil, ErrNotedUnavailable
	}

	// Build noted command
	args := []string{"add", content}

	if opts.Category != "" {
		args = append(args, "--category", opts.Category)
	}

	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}

	for key, value := range opts.Metadata {
		args = append(args, "--meta", fmt.Sprintf("%s=%s", key, value))
	}

	args = append(args, "--format", "json")

	cmd := exec.Command("noted", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("noted add failed: %w", err)
	}

	var memory Memory
	if err := json.Unmarshal(output, &memory); err != nil {
		// If JSON parsing fails, create a basic memory
		memory = Memory{
			Content:   content,
			Category:  opts.Category,
			Tags:      opts.Tags,
			Metadata:  opts.Metadata,
			CreatedAt: time.Now(),
		}
	}

	return &memory, nil
}

// RecallOptions configures memory recall
type RecallOptions struct {
	Query    string   // Search query
	Category string   // Filter by category
	Tags     []string // Filter by tags
	Limit    int      // Maximum results
	Since    string   // Only memories after this time (e.g., "1 week ago")
}

// Recall retrieves memories using noted
func (n *NotedProvider) Recall(opts RecallOptions) ([]Memory, error) {
	if !n.available {
		return nil, ErrNotedUnavailable
	}

	// Build noted search command
	args := []string{"search"}

	if opts.Query != "" {
		args = append(args, opts.Query)
	}

	if opts.Category != "" {
		args = append(args, "--category", opts.Category)
	}

	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}

	if opts.Limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", opts.Limit))
	}

	if opts.Since != "" {
		args = append(args, "--since", opts.Since)
	}

	args = append(args, "--format", "json")

	cmd := exec.Command("noted", args...)
	output, err := cmd.Output()
	if err != nil {
		// noted might return non-zero for no results
		if len(output) == 0 {
			return []Memory{}, nil
		}
		return nil, fmt.Errorf("noted search failed: %w", err)
	}

	var memories []Memory
	if err := json.Unmarshal(output, &memories); err != nil {
		// Try parsing as line-delimited JSON
		return parseLineDelimitedJSON(output)
	}

	return memories, nil
}

// parseLineDelimitedJSON parses newline-delimited JSON
func parseLineDelimitedJSON(data []byte) ([]Memory, error) {
	var memories []Memory
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var memory Memory
		if err := json.Unmarshal([]byte(line), &memory); err != nil {
			continue // Skip malformed lines
		}
		memories = append(memories, memory)
	}

	return memories, nil
}

// RecallContext retrieves contextually relevant memories for git work
func (n *NotedProvider) RecallContext(file string, branch string, recentCommits []string) ([]Memory, error) {
	if !n.available {
		return nil, ErrNotedUnavailable
	}

	var allMemories []Memory

	// Search by file path
	if file != "" {
		memories, err := n.Recall(RecallOptions{
			Query:    file,
			Category: "git",
			Limit:    5,
		})
		if err == nil {
			allMemories = append(allMemories, memories...)
		}
	}

	// Search by branch
	if branch != "" {
		memories, err := n.Recall(RecallOptions{
			Query:    branch,
			Category: "git",
			Limit:    3,
		})
		if err == nil {
			allMemories = append(allMemories, memories...)
		}
	}

	// Search by recent commit messages
	for _, commit := range recentCommits {
		if len(commit) > 50 {
			commit = commit[:50]
		}
		memories, err := n.Recall(RecallOptions{
			Query:    commit,
			Category: "git",
			Limit:    2,
		})
		if err == nil {
			allMemories = append(allMemories, memories...)
		}
	}

	// Deduplicate by ID
	seen := make(map[string]bool)
	var unique []Memory
	for _, m := range allMemories {
		if m.ID != "" && !seen[m.ID] {
			seen[m.ID] = true
			unique = append(unique, m)
		} else if m.ID == "" {
			unique = append(unique, m)
		}
	}

	// Limit total results
	if len(unique) > 10 {
		unique = unique[:10]
	}

	return unique, nil
}

// AutoCapture represents an event that should be auto-captured
type AutoCaptureEvent struct {
	Type        string            // "commit", "branch_create", "merge", etc.
	Description string            // Human-readable description
	Context     map[string]string // Additional context
}

// CaptureEvent automatically stores a git event as a memory
func (n *NotedProvider) CaptureEvent(event AutoCaptureEvent) error {
	if !n.available {
		return ErrNotedUnavailable
	}

	tags := []string{"auto-captured", event.Type}
	metadata := map[string]string{
		"event_type": event.Type,
		"source":     "gpeek",
	}

	for k, v := range event.Context {
		metadata[k] = v
	}

	_, err := n.Remember(event.Description, RememberOptions{
		Category: "git",
		Tags:     tags,
		Metadata: metadata,
	})

	return err
}
