package watcher

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWatcher_DetectsNewFile(t *testing.T) {
	dir := t.TempDir()

	w := New(dir)

	var mu sync.Mutex
	var events []string
	err := w.Start(func(path string) {
		mu.Lock()
		events = append(events, path)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer w.Stop()

	// Give watcher time to set up
	time.Sleep(100 * time.Millisecond)

	// Create a new .md file
	err = os.WriteFile(filepath.Join(dir, "new.md"), []byte("# New"), 0644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Wait for debounce + processing
	time.Sleep(600 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(events) == 0 {
		t.Error("expected at least one event for new.md, got none")
	}
	found := false
	for _, e := range events {
		if filepath.Base(e) == "new.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected event for new.md, got %v", events)
	}
}

func TestWatcher_DetectsModification(t *testing.T) {
	dir := t.TempDir()

	// Pre-create file
	mdPath := filepath.Join(dir, "exist.md")
	os.WriteFile(mdPath, []byte("# Old"), 0644)

	w := New(dir)
	var mu sync.Mutex
	var events []string
	err := w.Start(func(path string) {
		mu.Lock()
		events = append(events, path)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	// Modify the file
	os.WriteFile(mdPath, []byte("# Updated"), 0644)

	time.Sleep(600 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(events) == 0 {
		t.Error("expected at least one event for exist.md modification")
	}
}

func TestWatcher_DetectsDelete(t *testing.T) {
	dir := t.TempDir()

	// Pre-create file
	mdPath := filepath.Join(dir, "todelete.md")
	os.WriteFile(mdPath, []byte("# Delete me"), 0644)

	w := New(dir)
	var mu sync.Mutex
	var events []string
	err := w.Start(func(path string) {
		mu.Lock()
		events = append(events, path)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	os.Remove(mdPath)

	time.Sleep(600 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(events) == 0 {
		t.Error("expected at least one event for todelete.md removal")
	}
}

func TestWatcher_IgnoresNonMd(t *testing.T) {
	dir := t.TempDir()

	w := New(dir)
	var mu sync.Mutex
	var events []string
	err := w.Start(func(path string) {
		mu.Lock()
		events = append(events, path)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	// Create a .txt file — should be ignored
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("text"), 0644)

	time.Sleep(600 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(events) > 0 {
		t.Errorf("expected no events for .txt file, got %v", events)
	}
}

func TestWatcher_DetectsSubdirectory(t *testing.T) {
	dir := t.TempDir()

	w := New(dir)
	var mu sync.Mutex
	var events []string
	err := w.Start(func(path string) {
		mu.Lock()
		events = append(events, path)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	// Create a subdirectory and add a .md file
	subDir := filepath.Join(dir, "notes")
	os.Mkdir(subDir, 0755)

	time.Sleep(200 * time.Millisecond)

	os.WriteFile(filepath.Join(subDir, "sub.md"), []byte("# Sub"), 0644)

	time.Sleep(600 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	found := false
	for _, e := range events {
		if filepath.Base(e) == "sub.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected event for sub.md in subdirectory, got %v", events)
	}
}

func TestWatcher_Stop(t *testing.T) {
	dir := t.TempDir()

	w := New(dir)
	err := w.Start(func(path string) {})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Should not panic on stop
	w.Stop()

	// Should not panic on double stop
	w.Stop()
}
