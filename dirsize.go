package main

import (
	"os"
	"path/filepath"
	"sync"
)

// DirSizeCache provides async directory size calculation with caching.
type DirSizeCache struct {
	mu       sync.RWMutex
	sizes    map[string]int64
	pending  map[string]bool
	cancel   map[string]chan struct{}
	onUpdate func(path string, size int64)
}

// NewDirSizeCache creates a new directory size cache.
// onUpdate is called (from a goroutine) when a size calculation completes.
func NewDirSizeCache(onUpdate func(path string, size int64)) *DirSizeCache {
	return &DirSizeCache{
		sizes:    make(map[string]int64),
		pending:  make(map[string]bool),
		cancel:   make(map[string]chan struct{}),
		onUpdate: onUpdate,
	}
}

// Get returns the cached size for a path and whether it was found.
func (c *DirSizeCache) Get(path string) (int64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	size, ok := c.sizes[path]
	return size, ok
}

// IsPending returns true if a size calculation is in progress for the path.
func (c *DirSizeCache) IsPending(path string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pending[path]
}

// RequestSize starts an async size calculation for the given directory path.
// If already cached or pending, this is a no-op.
func (c *DirSizeCache) RequestSize(path string) {
	c.mu.Lock()
	if _, cached := c.sizes[path]; cached {
		c.mu.Unlock()
		return
	}
	if c.pending[path] {
		c.mu.Unlock()
		return
	}
	c.pending[path] = true
	cancelCh := make(chan struct{})
	c.cancel[path] = cancelCh
	c.mu.Unlock()

	go func() {
		size := calcDirSizeWithCancel(path, cancelCh)

		c.mu.Lock()
		// Only store if not invalidated while calculating
		if c.pending[path] {
			c.sizes[path] = size
			delete(c.pending, path)
			delete(c.cancel, path)
		}
		c.mu.Unlock()

		if c.onUpdate != nil {
			c.onUpdate(path, size)
		}
	}()
}

// RequestSizesForDir requests size calculations for all directory entries.
func (c *DirSizeCache) RequestSizesForDir(parentPath string, entries []FileEntry) {
	for _, e := range entries {
		if e.IsDir && e.Name != ".." && e.Accessible {
			c.RequestSize(filepath.Join(parentPath, e.Name))
		}
	}
}

// Invalidate removes a path from the cache and cancels any pending calculation.
func (c *DirSizeCache) Invalidate(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sizes, path)
	if ch, ok := c.cancel[path]; ok {
		close(ch)
		delete(c.cancel, path)
	}
	delete(c.pending, path)
}

// InvalidateAll clears the entire cache and cancels all pending calculations.
func (c *DirSizeCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for path, ch := range c.cancel {
		close(ch)
		delete(c.cancel, path)
	}
	c.sizes = make(map[string]int64)
	c.pending = make(map[string]bool)
}

// calcDirSizeWithCancel walks a directory tree summing file sizes.
// It checks the cancel channel periodically to allow early termination.
func calcDirSizeWithCancel(path string, cancel <-chan struct{}) int64 {
	var total int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		select {
		case <-cancel:
			return filepath.SkipAll
		default:
		}
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}
