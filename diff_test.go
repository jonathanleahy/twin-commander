package main

import (
	"strings"
	"testing"
)

func TestUnifiedDiff_Identical(t *testing.T) {
	lines := []string{"a", "b", "c"}
	result := unifiedDiff("left", "right", lines, lines)
	if !strings.Contains(result, "---") {
		// Header should still be present
		t.Error("expected diff header")
	}
	// No hunks for identical files
	if strings.Contains(result, "@@") {
		t.Error("expected no hunks for identical files")
	}
}

func TestUnifiedDiff_Addition(t *testing.T) {
	left := []string{"a", "b"}
	right := []string{"a", "b", "c"}
	result := unifiedDiff("left", "right", left, right)
	if !strings.Contains(result, "+c") {
		t.Errorf("expected +c in diff, got:\n%s", result)
	}
}

func TestUnifiedDiff_Deletion(t *testing.T) {
	left := []string{"a", "b", "c"}
	right := []string{"a", "c"}
	result := unifiedDiff("left", "right", left, right)
	if !strings.Contains(result, "-b") {
		t.Errorf("expected -b in diff, got:\n%s", result)
	}
}

func TestUnifiedDiff_Modification(t *testing.T) {
	left := []string{"a", "b", "c"}
	right := []string{"a", "x", "c"}
	result := unifiedDiff("left", "right", left, right)
	if !strings.Contains(result, "-b") || !strings.Contains(result, "+x") {
		t.Errorf("expected -b and +x in diff, got:\n%s", result)
	}
}

func TestIsBinaryContent(t *testing.T) {
	if isBinaryContent([]byte("hello world")) {
		t.Error("text should not be binary")
	}
	if !isBinaryContent([]byte("hello\x00world")) {
		t.Error("null byte should be binary")
	}
	if isBinaryContent([]byte{}) {
		t.Error("empty should not be binary")
	}
}
