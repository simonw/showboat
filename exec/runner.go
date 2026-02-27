package exec

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// runCmd runs cmd, capturing combined stdout+stderr. Non-zero exit codes are
// not treated as errors — the output is still returned alongside the exit code.
func runCmd(cmd *exec.Cmd, label string) (string, int, error) {
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return buf.String(), exitErr.ExitCode(), nil
		}
		return "", 1, fmt.Errorf("executing %s: %w", label, err)
	}

	return buf.String(), 0, nil
}

// Run executes code using the given language interpreter and returns
// the combined stdout+stderr output and the process exit code.
// Non-zero exit codes are not treated as errors — the output is still
// captured and returned alongside the exit code.
// If workdir is empty, the current directory is used.
func Run(lang, code, workdir string) (string, int, error) {
	cmd := exec.Command(lang, "-c", code)
	if workdir != "" {
		cmd.Dir = workdir
	}
	return runCmd(cmd, lang)
}

// RunWithFilter executes filter (a shell command string) with code written to
// its stdin, and returns the combined stdout+stderr output and exit code.
// The filter is run via "bash -c". If workdir is empty, the current directory
// is used.
func RunWithFilter(filter, code, workdir string) (string, int, error) {
	cmd := exec.Command("bash", "-c", filter)
	if workdir != "" {
		cmd.Dir = workdir
	}
	cmd.Stdin = strings.NewReader(code)
	return runCmd(cmd, "filter")
}
