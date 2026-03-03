package main

import (
	"bytes"
	"os/exec"
	"strings"
)

// CommandResult holds the output and status of an executed command.
type CommandResult struct {
	Output   string
	ExitCode int
	Err      error
}

// ParseCommand splits an input string into command name and arguments.
// Uses strings.Fields for simple splitting.
// Returns ("", nil) for empty input.
func ParseCommand(input string) (string, []string) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return "", nil
	}
	return fields[0], fields[1:]
}

// ExpandVariables replaces special variables in a command string:
// %f = selected file path, %d = current directory, %s = selected files (space-separated)
func ExpandVariables(input, selectedFile, currentDir string, selectedFiles []string) string {
	result := strings.ReplaceAll(input, "%f", selectedFile)
	result = strings.ReplaceAll(result, "%d", currentDir)
	result = strings.ReplaceAll(result, "%s", strings.Join(selectedFiles, " "))
	return result
}

// RunCommand executes a shell command in the given working directory.
// Returns CommandResult with stdout+stderr combined, exit code, and any error.
func RunCommand(cmdStr, workDir string) CommandResult {
	if strings.TrimSpace(cmdStr) == "" {
		return CommandResult{
			Output:   "",
			ExitCode: -1,
			Err:      &exec.Error{Name: "", Err: exec.ErrNotFound},
		}
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Dir = workDir

	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined

	err := cmd.Run()

	result := CommandResult{
		Output: combined.String(),
	}

	if err != nil {
		result.Err = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}
