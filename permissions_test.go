package main

import (
	"testing"
)

func TestFormatPermissions_755(t *testing.T) {
	got := FormatPermissions(0755)
	want := "rwxr-xr-x"
	if got != want {
		t.Errorf("FormatPermissions(0755) = %q, want %q", got, want)
	}
}

func TestFormatPermissions_644(t *testing.T) {
	got := FormatPermissions(0644)
	want := "rw-r--r--"
	if got != want {
		t.Errorf("FormatPermissions(0644) = %q, want %q", got, want)
	}
}

func TestFormatPermissions_Zero(t *testing.T) {
	got := FormatPermissions(0)
	want := "---------"
	if got != want {
		t.Errorf("FormatPermissions(0) = %q, want %q", got, want)
	}
}

func TestFormatPermissions_777(t *testing.T) {
	got := FormatPermissions(0777)
	want := "rwxrwxrwx"
	if got != want {
		t.Errorf("FormatPermissions(0777) = %q, want %q", got, want)
	}
}

func TestFormatPermissions_600(t *testing.T) {
	got := FormatPermissions(0600)
	want := "rw-------"
	if got != want {
		t.Errorf("FormatPermissions(0600) = %q, want %q", got, want)
	}
}

func TestParseOctalMode_755(t *testing.T) {
	mode, err := ParseOctalMode("755")
	if err != nil {
		t.Fatalf("ParseOctalMode(\"755\") returned error: %v", err)
	}
	if mode != 0o755 {
		t.Errorf("ParseOctalMode(\"755\") = %o, want 755", mode)
	}
}

func TestParseOctalMode_644(t *testing.T) {
	mode, err := ParseOctalMode("644")
	if err != nil {
		t.Fatalf("ParseOctalMode(\"644\") returned error: %v", err)
	}
	if mode != 0o644 {
		t.Errorf("ParseOctalMode(\"644\") = %o, want 644", mode)
	}
}

func TestParseOctalMode_4755(t *testing.T) {
	mode, err := ParseOctalMode("4755")
	if err != nil {
		t.Fatalf("ParseOctalMode(\"4755\") returned error: %v", err)
	}
	if mode != 0o4755 {
		t.Errorf("ParseOctalMode(\"4755\") = %o, want 4755", mode)
	}
}

func TestParseOctalMode_Error_999(t *testing.T) {
	_, err := ParseOctalMode("999")
	if err == nil {
		t.Error("ParseOctalMode(\"999\") expected error, got nil")
	}
}

func TestParseOctalMode_Error_Empty(t *testing.T) {
	_, err := ParseOctalMode("")
	if err == nil {
		t.Error("ParseOctalMode(\"\") expected error, got nil")
	}
}

func TestParseOctalMode_Error_Alpha(t *testing.T) {
	_, err := ParseOctalMode("abc")
	if err == nil {
		t.Error("ParseOctalMode(\"abc\") expected error, got nil")
	}
}

func TestParseOctalMode_Error_TooLong(t *testing.T) {
	_, err := ParseOctalMode("77777")
	if err == nil {
		t.Error("ParseOctalMode(\"77777\") expected error, got nil")
	}
}
