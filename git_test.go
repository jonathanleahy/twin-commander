package main

import "testing"

func TestGitStatusLabel(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"??", "?"},
		{"A ", "A"},
		{"M ", "M"},
		{" M", "M"},
		{"D ", "D"},
		{" D", "D"},
		{"R ", "R"},
		{"", ""},
		{"C ", "C"},
	}
	for _, tt := range tests {
		got := GitStatusLabel(tt.status)
		if got != tt.want {
			t.Errorf("GitStatusLabel(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}
