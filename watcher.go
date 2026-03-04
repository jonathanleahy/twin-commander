package main

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches directories for changes and triggers panel refreshes.
type FileWatcher struct {
	watcher    *fsnotify.Watcher
	app        *App
	done       chan struct{}
	mu         sync.Mutex
	watched    map[string]bool
	debounceMs int
}

// NewFileWatcher creates a file watcher attached to the given App.
func NewFileWatcher(app *App) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	fw := &FileWatcher{
		watcher:    w,
		app:        app,
		done:       make(chan struct{}),
		watched:    make(map[string]bool),
		debounceMs: 300,
	}
	go fw.loop()
	return fw, nil
}

// loop is the main event loop that listens for filesystem changes.
func (fw *FileWatcher) loop() {
	var timer *time.Timer
	for {
		select {
		case <-fw.done:
			if timer != nil {
				timer.Stop()
			}
			return
		case _, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			// Debounce: reset timer on each event
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(time.Duration(fw.debounceMs)*time.Millisecond, func() {
				fw.app.Application.QueueUpdateDraw(func() {
					fw.app.refreshAllPanels()
				})
			})
		case _, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			// Silently ignore watcher errors
		}
	}
}

// UpdateWatchedDirs watches the given directories and unwatches any previously
// watched directories that are no longer needed.
func (fw *FileWatcher) UpdateWatchedDirs(dirs []string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	wanted := make(map[string]bool)
	for _, d := range dirs {
		wanted[d] = true
	}

	// Remove directories no longer needed
	for d := range fw.watched {
		if !wanted[d] {
			_ = fw.watcher.Remove(d)
			delete(fw.watched, d)
		}
	}

	// Add new directories
	for d := range wanted {
		if !fw.watched[d] {
			if err := fw.watcher.Add(d); err == nil {
				fw.watched[d] = true
			}
		}
	}
}

// Close stops the watcher and releases resources.
func (fw *FileWatcher) Close() {
	close(fw.done)
	_ = fw.watcher.Close()
}
