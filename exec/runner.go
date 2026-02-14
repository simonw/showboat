package exec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

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

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return buf.String(), exitErr.ExitCode(), nil
		}
		return "", 1, fmt.Errorf("executing %s: %w", lang, err)
	}

	return buf.String(), 0, nil
}

// RunWithEnv is like Run but appends extra environment variables.
func RunWithEnv(lang, code, workdir string, env []string) (string, int, error) {
	cmd := exec.Command(lang, "-c", code)

	if workdir != "" {
		cmd.Dir = workdir
	}

	cmd.Env = append(os.Environ(), env...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return buf.String(), exitErr.ExitCode(), nil
		}
		return "", 1, fmt.Errorf("executing %s: %w", lang, err)
	}

	return buf.String(), 0, nil
}
