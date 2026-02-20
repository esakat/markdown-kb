package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// newTestRepo creates a temporary git repository with 2 commits
// affecting a test file, and returns the repo directory.
func newTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init", "-b", "main")

	// First commit
	os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# Hello\n\nFirst version.\n"), 0o644)
	run("add", "doc.md")
	run("commit", "-m", "initial: add doc.md")

	// Second commit
	os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# Hello\n\nUpdated version.\nNew line added.\n"), 0o644)
	run("add", "doc.md")
	run("commit", "-m", "update: modify doc.md")

	// Add a file in subdirectory
	os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "sub", "nested.md"), []byte("# Nested\n"), 0o644)
	run("add", "sub/nested.md")
	run("commit", "-m", "add: sub/nested.md")

	return dir
}

func TestFileHistory(t *testing.T) {
	dir := newTestRepo(t)

	commits, err := FileHistory(dir, "doc.md")
	if err != nil {
		t.Fatalf("FileHistory() error = %v", err)
	}
	if len(commits) < 2 {
		t.Fatalf("expected at least 2 commits, got %d", len(commits))
	}

	// Most recent first
	if !strings.Contains(commits[0].Message, "update") {
		t.Errorf("first commit message = %q, expected to contain 'update'", commits[0].Message)
	}
	if !strings.Contains(commits[1].Message, "initial") {
		t.Errorf("second commit message = %q, expected to contain 'initial'", commits[1].Message)
	}

	// Check fields
	for _, c := range commits {
		if c.Hash == "" {
			t.Error("expected non-empty hash")
		}
		if c.Author == "" {
			t.Error("expected non-empty author")
		}
		if c.Date.IsZero() {
			t.Error("expected non-zero date")
		}
	}
}

func TestFileHistory_SubdirectoryFile(t *testing.T) {
	dir := newTestRepo(t)

	commits, err := FileHistory(dir, "sub/nested.md")
	if err != nil {
		t.Fatalf("FileHistory() error = %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("expected 1 commit for sub/nested.md, got %d", len(commits))
	}
}

func TestFileHistory_NonexistentFile(t *testing.T) {
	dir := newTestRepo(t)

	commits, err := FileHistory(dir, "nonexistent.md")
	if err != nil {
		t.Fatalf("FileHistory() error = %v", err)
	}
	if len(commits) != 0 {
		t.Errorf("expected 0 commits for nonexistent file, got %d", len(commits))
	}
}

func TestDiff(t *testing.T) {
	dir := newTestRepo(t)

	commits, err := FileHistory(dir, "doc.md")
	if err != nil {
		t.Fatalf("FileHistory() error = %v", err)
	}
	if len(commits) < 2 {
		t.Fatal("need at least 2 commits")
	}

	oldHash := commits[1].Hash // initial
	newHash := commits[0].Hash // update

	diff, err := Diff(dir, "doc.md", oldHash, newHash)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	if diff == "" {
		t.Error("expected non-empty diff")
	}
	if !strings.Contains(diff, "Updated version") {
		t.Errorf("diff should contain 'Updated version', got:\n%s", diff)
	}
	if !strings.Contains(diff, "First version") {
		t.Errorf("diff should contain 'First version', got:\n%s", diff)
	}
}

func TestBlame(t *testing.T) {
	dir := newTestRepo(t)

	lines, err := Blame(dir, "doc.md")
	if err != nil {
		t.Fatalf("Blame() error = %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("expected non-empty blame result")
	}

	for _, line := range lines {
		if line.Hash == "" {
			t.Error("expected non-empty hash")
		}
		if line.Author == "" {
			t.Error("expected non-empty author")
		}
		if line.LineNo < 1 {
			t.Errorf("expected positive line number, got %d", line.LineNo)
		}
	}
}

func TestBlameRange(t *testing.T) {
	dir := newTestRepo(t)

	lines, err := BlameRange(dir, "doc.md", 1, 2)
	if err != nil {
		t.Fatalf("BlameRange() error = %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0].LineNo != 1 {
		t.Errorf("first line number = %d, want 1", lines[0].LineNo)
	}
	if lines[1].LineNo != 2 {
		t.Errorf("second line number = %d, want 2", lines[1].LineNo)
	}
}

func TestFileDates(t *testing.T) {
	dir := newTestRepo(t)

	created, updated, err := FileDates(dir, "doc.md")
	if err != nil {
		t.Fatalf("FileDates() error = %v", err)
	}
	if created.IsZero() {
		t.Error("expected non-zero created date")
	}
	if updated.IsZero() {
		t.Error("expected non-zero updated date")
	}
	if !updated.After(created) && !updated.Equal(created) {
		t.Errorf("updated (%v) should be >= created (%v)", updated, created)
	}
}

func TestFileDates_SingleCommit(t *testing.T) {
	dir := newTestRepo(t)

	created, updated, err := FileDates(dir, "sub/nested.md")
	if err != nil {
		t.Fatalf("FileDates() error = %v", err)
	}
	if created.IsZero() || updated.IsZero() {
		t.Error("expected non-zero dates")
	}
	// For single commit, created == updated
	if !created.Equal(updated) {
		t.Errorf("expected created == updated for single-commit file, got %v vs %v", created, updated)
	}
}

func TestFileHistory_NotGitRepo(t *testing.T) {
	dir := t.TempDir() // not a git repo

	_, err := FileHistory(dir, "anything.md")
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}
