package git

import (
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

func createTestRepo(t *testing.T) (*Repository, string) {
	t.Helper()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gpeek-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Initialize a git repository
	repo, err := gogit.PlainInit(tmpDir, false)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create a Repository wrapper
	r, err := Open(tmpDir)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to open repo: %v", err)
	}

	// Store the underlying repo for modifications
	r.repo = repo

	return r, tmpDir
}

func TestRepository_DefaultRemote(t *testing.T) {
	r, tmpDir := createTestRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Without any remotes, should return "origin"
	if got := r.DefaultRemote(); got != "origin" {
		t.Errorf("DefaultRemote() = %q, want %q", got, "origin")
	}

	// Set a custom remote
	r.SetDefaultRemote("upstream")
	if got := r.DefaultRemote(); got != "upstream" {
		t.Errorf("DefaultRemote() = %q, want %q", got, "upstream")
	}
}

func TestRepository_DetectDefaultRemote_WithOrigin(t *testing.T) {
	r, tmpDir := createTestRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add origin remote
	_, err := r.repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/example/repo.git"},
	})
	if err != nil {
		t.Fatalf("failed to create remote: %v", err)
	}

	// Detect should prefer origin
	detected := r.detectDefaultRemote()
	if detected != "origin" {
		t.Errorf("detectDefaultRemote() = %q, want %q", detected, "origin")
	}
}

func TestRepository_DetectDefaultRemote_WithoutOrigin(t *testing.T) {
	r, tmpDir := createTestRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add a non-origin remote
	_, err := r.repo.CreateRemote(&config.RemoteConfig{
		Name: "upstream",
		URLs: []string{"https://github.com/example/repo.git"},
	})
	if err != nil {
		t.Fatalf("failed to create remote: %v", err)
	}

	// Detect should return first remote
	detected := r.detectDefaultRemote()
	if detected != "upstream" {
		t.Errorf("detectDefaultRemote() = %q, want %q", detected, "upstream")
	}
}

func TestRepository_DetectDefaultRemote_PreferOrigin(t *testing.T) {
	r, tmpDir := createTestRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add multiple remotes, origin not first
	_, err := r.repo.CreateRemote(&config.RemoteConfig{
		Name: "upstream",
		URLs: []string{"https://github.com/other/repo.git"},
	})
	if err != nil {
		t.Fatalf("failed to create upstream remote: %v", err)
	}

	_, err = r.repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/example/repo.git"},
	})
	if err != nil {
		t.Fatalf("failed to create origin remote: %v", err)
	}

	// Detect should still prefer origin
	detected := r.detectDefaultRemote()
	if detected != "origin" {
		t.Errorf("detectDefaultRemote() = %q, want %q", detected, "origin")
	}
}

func TestRepository_GetRemotes(t *testing.T) {
	r, tmpDir := createTestRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Initially no remotes
	remotes, err := r.GetRemotes()
	if err != nil {
		t.Fatalf("GetRemotes() error = %v", err)
	}
	if len(remotes) != 0 {
		t.Errorf("expected 0 remotes, got %d", len(remotes))
	}

	// Add a remote
	err = r.AddRemote("origin", "https://github.com/example/repo.git")
	if err != nil {
		t.Fatalf("AddRemote() error = %v", err)
	}

	remotes, err = r.GetRemotes()
	if err != nil {
		t.Fatalf("GetRemotes() error = %v", err)
	}
	if len(remotes) != 1 {
		t.Errorf("expected 1 remote, got %d", len(remotes))
	}
	if remotes[0] != "origin" {
		t.Errorf("expected remote name 'origin', got %q", remotes[0])
	}
}

func TestRepository_RemoveRemote(t *testing.T) {
	r, tmpDir := createTestRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Add and remove a remote
	err := r.AddRemote("origin", "https://github.com/example/repo.git")
	if err != nil {
		t.Fatalf("AddRemote() error = %v", err)
	}

	err = r.RemoveRemote("origin")
	if err != nil {
		t.Fatalf("RemoveRemote() error = %v", err)
	}

	remotes, _ := r.GetRemotes()
	if len(remotes) != 0 {
		t.Errorf("expected 0 remotes after removal, got %d", len(remotes))
	}
}

func TestRepository_Open_DetectsDefaultRemote(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "gpeek-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Initialize a git repository
	repo, err := gogit.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Add a remote before opening
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "myremote",
		URLs: []string{"https://example.com/repo.git"},
	})
	if err != nil {
		t.Fatalf("failed to create remote: %v", err)
	}

	// Open the repository - should auto-detect remote
	r, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	// Should have detected "myremote" as default since there's no origin
	if r.DefaultRemote() != "myremote" {
		t.Errorf("DefaultRemote() = %q, want %q", r.DefaultRemote(), "myremote")
	}
}

func TestRepository_Path(t *testing.T) {
	r, tmpDir := createTestRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	absPath, _ := filepath.Abs(tmpDir)
	if r.Path() != absPath {
		t.Errorf("Path() = %q, want %q", r.Path(), absPath)
	}
}

func TestRepository_IsValid(t *testing.T) {
	r, tmpDir := createTestRepo(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if !r.IsValid() {
		t.Error("IsValid() = false, want true")
	}
}
