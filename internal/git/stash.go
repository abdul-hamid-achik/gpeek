package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Stash struct {
	Index   int       `json:"index"`
	Message string    `json:"message"`
	Branch  string    `json:"branch,omitempty"`
	Hash    string    `json:"hash"`
	Time    time.Time `json:"time"`
}

// StashSave saves the current changes to the stash with an optional message
func (r *Repository) StashSave(message string) error {
	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stash save failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// StashList returns a list of all stashes
func (r *Repository) StashList() ([]Stash, error) {
	cmd := exec.Command("git", "stash", "list", "--format=%gd|%s|%H|%ai")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		// No stashes is not an error
		return nil, nil
	}

	var stashes []Stash
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	// Pattern to match stash@{n}
	stashPattern := regexp.MustCompile(`stash@\{(\d+)\}`)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 4 {
			continue
		}

		stashRef := parts[0]
		message := parts[1]
		hash := parts[2]
		timeStr := parts[3]

		// Extract index from stash@{n}
		matches := stashPattern.FindStringSubmatch(stashRef)
		if len(matches) < 2 {
			continue
		}
		index, _ := strconv.Atoi(matches[1])

		// Parse branch from message like "WIP on main: abc1234 commit message"
		// or "On main: message"
		branch := ""
		if strings.HasPrefix(message, "WIP on ") || strings.HasPrefix(message, "On ") {
			colonIdx := strings.Index(message, ":")
			if colonIdx > 0 {
				prefix := message[:colonIdx]
				if strings.HasPrefix(prefix, "WIP on ") {
					branch = strings.TrimPrefix(prefix, "WIP on ")
				} else if strings.HasPrefix(prefix, "On ") {
					branch = strings.TrimPrefix(prefix, "On ")
				}
			}
		}

		// Parse time
		parsedTime, _ := time.Parse("2006-01-02 15:04:05 -0700", timeStr)

		stashes = append(stashes, Stash{
			Index:   index,
			Message: message,
			Branch:  branch,
			Hash:    hash,
			Time:    parsedTime,
		})
	}

	return stashes, nil
}

// StashPop applies and removes the stash at the given index
func (r *Repository) StashPop(index int) error {
	cmd := exec.Command("git", "stash", "pop", fmt.Sprintf("stash@{%d}", index))
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stash pop failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// StashApply applies the stash at the given index without removing it
func (r *Repository) StashApply(index int) error {
	cmd := exec.Command("git", "stash", "apply", fmt.Sprintf("stash@{%d}", index))
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stash apply failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// StashDrop removes the stash at the given index
func (r *Repository) StashDrop(index int) error {
	cmd := exec.Command("git", "stash", "drop", fmt.Sprintf("stash@{%d}", index))
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stash drop failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// StashShow returns the diff for the stash at the given index
func (r *Repository) StashShow(index int) (string, error) {
	cmd := exec.Command("git", "stash", "show", "-p", fmt.Sprintf("stash@{%d}", index))
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("stash show failed: %v", err)
	}
	return string(output), nil
}
