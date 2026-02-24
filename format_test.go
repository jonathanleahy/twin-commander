package main

import (
	"testing"
)

// TS-19: File Size Formatting -- Bytes
func TestFormatSize_Bytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{450, "450"},
		{1023, "1023"},
	}
	for _, tc := range tests {
		result := FormatSize(tc.input)
		if result != tc.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// TS-20: File Size Formatting -- Kilobytes
func TestFormatSize_Kilobytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{1024, "1.0K"},
		{4096, "4.0K"},
		{46285, "45.2K"},
	}
	for _, tc := range tests {
		result := FormatSize(tc.input)
		if result != tc.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// TS-21: File Size Formatting -- Megabytes and Gigabytes
func TestFormatSize_MegabytesAndGigabytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{1048576, "1.0M"},
		{1363149, "1.3M"},
		{2254857830, "2.1G"},
	}
	for _, tc := range tests {
		result := FormatSize(tc.input)
		if result != tc.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// EC-5: Size Boundary Values
func TestFormatSize_BoundaryValues(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{1023, "1023"},
		{1024, "1.0K"},
		{1025, "1.0K"},
		{1048575, "1024.0K"},
		{1048576, "1.0M"},
		{1073741823, "1024.0M"},
		{1073741824, "1.0G"},
	}
	for _, tc := range tests {
		result := FormatSize(tc.input)
		if result != tc.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// EC-6: Zero-Byte File
func TestFormatSize_ZeroByte(t *testing.T) {
	result := FormatSize(0)
	if result != "0" {
		t.Errorf("FormatSize(0) = %q, want %q", result, "0")
	}
}

// EC-7: Very Large File (> 1 GB)
func TestFormatSize_VeryLargeFiles(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{1073741824, "1.0G"},
		{5368709120, "5.0G"},
	}
	for _, tc := range tests {
		result := FormatSize(tc.input)
		if result != tc.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// EC-18: FormatSize with Inaccessible Sentinel
func TestFormatSize_InaccessibleSentinel(t *testing.T) {
	result := FormatSize(-1)
	if result != "---" {
		t.Errorf("FormatSize(-1) = %q, want %q", result, "---")
	}
}
