package main

import (
	"strings"
	"testing"
)

func TestParseCommand_Simple(t *testing.T) {
	cmd, args := ParseCommand("ls")
	if cmd != "ls" {
		t.Errorf("command = %q, want %q", cmd, "ls")
	}
	if len(args) != 0 {
		t.Errorf("args = %v, want empty slice", args)
	}
}

func TestParseCommand_WithArgs(t *testing.T) {
	cmd, args := ParseCommand("ls -la /tmp")
	if cmd != "ls" {
		t.Errorf("command = %q, want %q", cmd, "ls")
	}
	if len(args) != 2 || args[0] != "-la" || args[1] != "/tmp" {
		t.Errorf("args = %v, want [-la /tmp]", args)
	}
}

func TestParseCommand_Empty(t *testing.T) {
	cmd, args := ParseCommand("")
	if cmd != "" {
		t.Errorf("command = %q, want empty string", cmd)
	}
	if args != nil {
		t.Errorf("args = %v, want nil", args)
	}
}

func TestParseCommand_WhitespaceOnly(t *testing.T) {
	cmd, args := ParseCommand("   ")
	if cmd != "" {
		t.Errorf("command = %q, want empty string", cmd)
	}
	if args != nil {
		t.Errorf("args = %v, want nil", args)
	}
}

func TestExpandVariables_File(t *testing.T) {
	result := ExpandVariables("cat %f", "/home/user/test.txt", "/home/user", nil)
	expected := "cat /home/user/test.txt"
	if result != expected {
		t.Errorf("result = %q, want %q", result, expected)
	}
}

func TestExpandVariables_Dir(t *testing.T) {
	result := ExpandVariables("ls %d", "", "/home/user", nil)
	expected := "ls /home/user"
	if result != expected {
		t.Errorf("result = %q, want %q", result, expected)
	}
}

func TestExpandVariables_Selected(t *testing.T) {
	files := []string{"a.txt", "b.txt", "c.txt"}
	result := ExpandVariables("rm %s", "", "/home/user", files)
	expected := "rm a.txt b.txt c.txt"
	if result != expected {
		t.Errorf("result = %q, want %q", result, expected)
	}
}

func TestExpandVariables_NoVars(t *testing.T) {
	result := ExpandVariables("echo hello", "/home/user/test.txt", "/home/user", []string{"a.txt"})
	expected := "echo hello"
	if result != expected {
		t.Errorf("result = %q, want %q", result, expected)
	}
}

func TestRunCommand_Success(t *testing.T) {
	result := RunCommand("echo hello", "/tmp")
	if result.Output != "hello\n" {
		t.Errorf("output = %q, want %q", result.Output, "hello\n")
	}
	if result.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", result.ExitCode)
	}
	if result.Err != nil {
		t.Errorf("err = %v, want nil", result.Err)
	}
}

func TestRunCommand_Failure(t *testing.T) {
	result := RunCommand("false", "/tmp")
	if result.ExitCode != 1 {
		t.Errorf("exit code = %d, want 1", result.ExitCode)
	}
	if result.Err == nil {
		t.Error("err should not be nil for failed command")
	}
}

func TestRunCommand_WorkDir(t *testing.T) {
	result := RunCommand("pwd", "/tmp")
	if !strings.Contains(result.Output, "/tmp") {
		t.Errorf("output = %q, want it to contain /tmp", result.Output)
	}
	if result.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", result.ExitCode)
	}
}

func TestRunCommand_Empty(t *testing.T) {
	result := RunCommand("", "/tmp")
	if result.Err == nil {
		t.Error("err should not be nil for empty command")
	}
}
