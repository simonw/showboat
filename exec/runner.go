package exec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Run executes code using the given language interpreter and returns
// the combined stdout+stderr output and the process exit code.
// Non-zero exit codes are not treated as errors — the output is still
// captured and returned alongside the exit code.
// If workdir is empty, the current directory is used.
// If varsFile is non-empty, the child process gets SHOWBOAT_VARS set
// and the showboat binary is made available on PATH.
func Run(lang, code, workdir, varsFile string) (string, int, error) {
	cmd := exec.Command(lang, "-c", code)

	if workdir != "" {
		cmd.Dir = workdir
	}

	if varsFile != "" {
		env := os.Environ()
		env = append(env, "SHOWBOAT_VARS="+varsFile)

		dir, cleanup := showboatDir()
		if cleanup != nil {
			defer cleanup()
		}
		if dir != "" {
			pathSep := ":"
			if runtime.GOOS == "windows" {
				pathSep = ";"
			}
			prependToPath(env, dir, pathSep)
		}

		cmd.Env = env
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

// showboatDir returns a directory containing a "showboat" binary suitable
// for prepending to PATH. It tries the cheapest option first:
//
//  1. If the directory containing os.Executable() already has a file named
//     "showboat" (or "showboat.exe" on Windows), use that directory directly
//     — no copy, no symlink, no cleanup.
//  2. Otherwise, create a temp directory with a symlink to the binary.
//     On Windows (where symlinks need privileges), fall back to a copy.
//
// Returns ("", nil) if the binary cannot be located.
func showboatDir() (dir string, cleanup func()) {
	self, err := os.Executable()
	if err != nil {
		return "", nil
	}
	return showboatDirFrom(self)
}

// showboatDirFrom is the testable core of showboatDir. Given the path to
// the current binary, it returns a directory containing a "showboat" entry.
func showboatDirFrom(self string) (dir string, cleanup func()) {
	name := "showboat"
	if runtime.GOOS == "windows" {
		name = "showboat.exe"
	}

	// Fast path: the binary's own directory already contains "showboat".
	// Covers: go install, system PATH, direct ./showboat invocation.
	selfDir := filepath.Dir(self)
	if _, err := os.Stat(filepath.Join(selfDir, name)); err == nil {
		return selfDir, nil
	}

	// Slow path: create a temp dir with a link to the real binary.
	// Covers: uvx (Python shim), go run, renamed binaries.
	tmpDir, err := os.MkdirTemp("", "showboat-path-*")
	if err != nil {
		return "", nil
	}
	dest := filepath.Join(tmpDir, name)

	// Try symlink first (free, works on Linux/macOS).
	if os.Symlink(self, dest) == nil {
		return tmpDir, func() { os.RemoveAll(tmpDir) }
	}

	// Symlink failed (Windows without dev mode) — fall back to copy.
	if copyBinary(self, dest) == nil {
		return tmpDir, func() { os.RemoveAll(tmpDir) }
	}

	// Both failed; clean up and give up.
	os.RemoveAll(tmpDir)
	return "", nil
}

// prependToPath adds dir to the front of the PATH entry in env (in-place).
func prependToPath(env []string, dir, sep string) {
	for i, e := range env {
		if idx := strings.IndexByte(e, '='); idx > 0 {
			key := e[:idx]
			if strings.EqualFold(key, "PATH") {
				env[i] = key + "=" + dir + sep + e[idx+1:]
				return
			}
		}
	}
	// No PATH found; add one
	env = append(env, "PATH="+dir)
}

// copyBinary copies a file from src to dst with executable permissions.
func copyBinary(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
