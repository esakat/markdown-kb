package watcher

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// skipDirs are directories that should not be watched.
var skipDirs = map[string]bool{
	".git":         true,
	".svn":         true,
	".hg":          true,
	"node_modules": true,
}

// Watcher monitors the file system for changes to Markdown files
// and triggers re-indexing.
type Watcher struct {
	rootDir  string
	fsw      *fsnotify.Watcher
	done     chan struct{}
	stopped  bool
	stopOnce sync.Once
}

// New creates a new file watcher for the given directory.
func New(rootDir string) *Watcher {
	return &Watcher{
		rootDir: rootDir,
		done:    make(chan struct{}),
	}
}

// Start begins watching for file changes. onChange is called with the
// relative path of the changed .md file. Events are debounced per file.
func (w *Watcher) Start(onChange func(path string)) error {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.fsw = fsw

	// Add root and all subdirectories
	if err := w.addDirs(w.rootDir); err != nil {
		fsw.Close()
		return err
	}

	go w.loop(onChange)

	return nil
}

// Stop terminates the file watcher.
func (w *Watcher) Stop() {
	w.stopOnce.Do(func() {
		close(w.done)
		if w.fsw != nil {
			w.fsw.Close()
		}
	})
}

func (w *Watcher) addDirs(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && (skipDirs[name] || strings.HasPrefix(name, ".")) {
				return fs.SkipDir
			}
			return w.fsw.Add(path)
		}
		return nil
	})
}

func (w *Watcher) loop(onChange func(path string)) {
	// Debounce: collect events per file, fire after 300ms of silence
	const debounce = 300 * time.Millisecond
	pending := make(map[string]*time.Timer)
	var mu sync.Mutex

	fire := func(path string) {
		mu.Lock()
		delete(pending, path)
		mu.Unlock()

		rel, err := filepath.Rel(w.rootDir, path)
		if err != nil {
			rel = path
		}
		onChange(rel)
	}

	for {
		select {
		case <-w.done:
			mu.Lock()
			for _, t := range pending {
				t.Stop()
			}
			mu.Unlock()
			return

		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}

			path := event.Name

			// If a new directory is created, watch it recursively
			if event.Has(fsnotify.Create) {
				if info, err := os.Stat(path); err == nil && info.IsDir() {
					w.addDirs(path)
					continue
				}
			}

			// Only care about .md files
			if filepath.Ext(path) != ".md" {
				continue
			}

			// Debounce per file
			mu.Lock()
			if t, ok := pending[path]; ok {
				t.Reset(debounce)
			} else {
				p := path
				pending[path] = time.AfterFunc(debounce, func() { fire(p) })
			}
			mu.Unlock()

		case _, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
		}
	}
}
