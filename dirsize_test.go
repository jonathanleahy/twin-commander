package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDirSizeCache_BasicCalculation(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), make([]byte, 100), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), make([]byte, 200), 0644)

	var wg sync.WaitGroup
	wg.Add(1)
	var gotSize int64

	cache := NewDirSizeCache(func(path string, size int64) {
		gotSize = size
		wg.Done()
	})

	cache.RequestSize(dir)
	wg.Wait()

	if gotSize != 300 {
		t.Errorf("expected size 300, got %d", gotSize)
	}

	// Should be cached now
	size, ok := cache.Get(dir)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if size != 300 {
		t.Errorf("cached size: expected 300, got %d", size)
	}
}

func TestDirSizeCache_NestedDirectories(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	os.MkdirAll(subdir, 0755)
	os.WriteFile(filepath.Join(dir, "top.txt"), make([]byte, 50), 0644)
	os.WriteFile(filepath.Join(subdir, "nested.txt"), make([]byte, 150), 0644)

	var wg sync.WaitGroup
	wg.Add(1)
	var gotSize int64

	cache := NewDirSizeCache(func(path string, size int64) {
		gotSize = size
		wg.Done()
	})

	cache.RequestSize(dir)
	wg.Wait()

	if gotSize != 200 {
		t.Errorf("expected size 200 (50+150), got %d", gotSize)
	}
}

func TestDirSizeCache_CacheHit(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), make([]byte, 42), 0644)

	var wg sync.WaitGroup
	wg.Add(1)
	callCount := 0

	cache := NewDirSizeCache(func(path string, size int64) {
		callCount++
		wg.Done()
	})

	cache.RequestSize(dir)
	wg.Wait()

	// Second request should be a no-op (cached)
	cache.RequestSize(dir)
	time.Sleep(50 * time.Millisecond)

	if callCount != 1 {
		t.Errorf("expected onUpdate called once, got %d", callCount)
	}
}

func TestDirSizeCache_Invalidation(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), make([]byte, 100), 0644)

	var wg sync.WaitGroup
	wg.Add(1)

	cache := NewDirSizeCache(func(path string, size int64) {
		wg.Done()
	})

	cache.RequestSize(dir)
	wg.Wait()

	// Should be cached
	_, ok := cache.Get(dir)
	if !ok {
		t.Fatal("expected cache hit before invalidation")
	}

	// Invalidate
	cache.Invalidate(dir)
	_, ok = cache.Get(dir)
	if ok {
		t.Error("expected cache miss after invalidation")
	}

	// Should be able to recalculate
	wg.Add(1)
	cache.RequestSize(dir)
	wg.Wait()

	_, ok = cache.Get(dir)
	if !ok {
		t.Error("expected cache hit after recalculation")
	}
}

func TestDirSizeCache_InvalidateAll(t *testing.T) {
	dir := t.TempDir()
	sub1 := filepath.Join(dir, "a")
	sub2 := filepath.Join(dir, "b")
	os.MkdirAll(sub1, 0755)
	os.MkdirAll(sub2, 0755)
	os.WriteFile(filepath.Join(sub1, "f.txt"), make([]byte, 10), 0644)
	os.WriteFile(filepath.Join(sub2, "g.txt"), make([]byte, 20), 0644)

	var wg sync.WaitGroup
	wg.Add(2)

	cache := NewDirSizeCache(func(path string, size int64) {
		wg.Done()
	})

	cache.RequestSize(sub1)
	cache.RequestSize(sub2)
	wg.Wait()

	cache.InvalidateAll()
	_, ok1 := cache.Get(sub1)
	_, ok2 := cache.Get(sub2)
	if ok1 || ok2 {
		t.Error("expected all cache entries cleared after InvalidateAll")
	}
}

func TestDirSizeCache_Cancel(t *testing.T) {
	// Create a dir with some files
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), make([]byte, 100), 0644)

	cache := NewDirSizeCache(func(path string, size int64) {})

	cache.RequestSize(dir)
	// Immediately invalidate to trigger cancel
	cache.Invalidate(dir)

	// Should not be pending or cached
	time.Sleep(50 * time.Millisecond)
	if cache.IsPending(dir) {
		t.Error("should not be pending after invalidation")
	}
}

func TestDirSizeCache_RequestSizesForDir(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "adir"), 0755)
	os.MkdirAll(filepath.Join(dir, "bdir"), 0755)
	os.WriteFile(filepath.Join(dir, "file.txt"), make([]byte, 10), 0644)
	os.WriteFile(filepath.Join(dir, "adir", "x.txt"), make([]byte, 50), 0644)
	os.WriteFile(filepath.Join(dir, "bdir", "y.txt"), make([]byte, 75), 0644)

	var wg sync.WaitGroup
	wg.Add(2)

	cache := NewDirSizeCache(func(path string, size int64) {
		wg.Done()
	})

	entries := []FileEntry{
		{Name: "..", IsDir: true, Accessible: true},
		{Name: "adir", IsDir: true, Accessible: true},
		{Name: "bdir", IsDir: true, Accessible: true},
		{Name: "file.txt", IsDir: false, Accessible: true},
	}

	cache.RequestSizesForDir(dir, entries)
	wg.Wait()

	aSize, ok := cache.Get(filepath.Join(dir, "adir"))
	if !ok || aSize != 50 {
		t.Errorf("adir: expected 50, got %d (ok=%v)", aSize, ok)
	}

	bSize, ok := cache.Get(filepath.Join(dir, "bdir"))
	if !ok || bSize != 75 {
		t.Errorf("bdir: expected 75, got %d (ok=%v)", bSize, ok)
	}
}
