package main

import (
	"testing"
	"time"
)

func TestCompareDirs_Identical(t *testing.T) {
	now := time.Now()
	left := []FileEntry{{Name: "a.txt", Size: 100, ModTime: now}}
	right := []FileEntry{{Name: "a.txt", Size: 100, ModTime: now}}

	diffs := CompareDirs(left, right)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestCompareDirs_LeftOnly(t *testing.T) {
	left := []FileEntry{{Name: "a.txt", Size: 100}}
	right := []FileEntry{}

	diffs := CompareDirs(left, right)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Kind != DiffLeftOnly {
		t.Errorf("expected DiffLeftOnly, got %v", diffs[0].Kind)
	}
}

func TestCompareDirs_RightOnly(t *testing.T) {
	left := []FileEntry{}
	right := []FileEntry{{Name: "b.txt", Size: 200}}

	diffs := CompareDirs(left, right)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Kind != DiffRightOnly {
		t.Errorf("expected DiffRightOnly, got %v", diffs[0].Kind)
	}
}

func TestCompareDirs_SizeDiff(t *testing.T) {
	now := time.Now()
	left := []FileEntry{{Name: "a.txt", Size: 100, ModTime: now}}
	right := []FileEntry{{Name: "a.txt", Size: 200, ModTime: now}}

	diffs := CompareDirs(left, right)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Kind != DiffSizeDiff {
		t.Errorf("expected DiffSizeDiff, got %v", diffs[0].Kind)
	}
}

func TestCompareDirs_DateDiff(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(time.Hour)
	left := []FileEntry{{Name: "a.txt", Size: 100, ModTime: t1}}
	right := []FileEntry{{Name: "a.txt", Size: 100, ModTime: t2}}

	diffs := CompareDirs(left, right)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Kind != DiffDateDiff {
		t.Errorf("expected DiffDateDiff, got %v", diffs[0].Kind)
	}
}

func TestCompareDirs_SkipsDotDot(t *testing.T) {
	left := []FileEntry{{Name: ".."}}
	right := []FileEntry{}

	diffs := CompareDirs(left, right)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs (.. skipped), got %d", len(diffs))
	}
}

func TestFormatDirComparison_Empty(t *testing.T) {
	result := FormatDirComparison(nil, "/a", "/b")
	if result != "Directories are identical." {
		t.Errorf("unexpected result: %s", result)
	}
}
