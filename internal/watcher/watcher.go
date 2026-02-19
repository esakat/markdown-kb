package watcher

// Watcher monitors the file system for changes to Markdown files
// and triggers re-indexing.
type Watcher struct {
	rootDir string
	done    chan struct{}
}

// New creates a new file watcher for the given directory.
func New(rootDir string) *Watcher {
	return &Watcher{
		rootDir: rootDir,
		done:    make(chan struct{}),
	}
}

// Start begins watching for file changes.
func (w *Watcher) Start(onChange func(path string)) error {
	// TODO: implement using fsnotify
	_ = onChange
	return nil
}

// Stop terminates the file watcher.
func (w *Watcher) Stop() {
	close(w.done)
}
