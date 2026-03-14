package exec

import (
	"strings"
	"testing"
)

func TestRunBash(t *testing.T) {
	output, _, err := Run("bash", "echo hello", "")
	if err != nil {
		t.Fatal(err)
	}
	if output != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", output)
	}
}

func TestRunPython(t *testing.T) {
	output, _, err := Run("python3", "print('hi')", "")
	if err != nil {
		t.Fatal(err)
	}
	if output != "hi\n" {
		t.Errorf("expected 'hi\\n', got %q", output)
	}
}

func TestRunWithWorkdir(t *testing.T) {
	output, _, err := Run("bash", "pwd", "/tmp")
	if err != nil {
		t.Fatal(err)
	}
	if output != "/tmp\n" {
		t.Errorf("expected '/tmp\\n', got %q", output)
	}
}

func TestRunNonZeroExit(t *testing.T) {
	output, exitCode, err := Run("bash", "echo oops && exit 1", "")
	if err != nil {
		t.Fatal(err)
	}
	if output != "oops\n" {
		t.Errorf("expected 'oops\\n', got %q", output)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestRunExitCodeReflected(t *testing.T) {
	_, exitCode, err := Run("bash", "exit 42", "")
	if err != nil {
		t.Fatal(err)
	}
	if exitCode != 42 {
		t.Errorf("expected exit code 42, got %d", exitCode)
	}
}

func TestRunZeroExitCode(t *testing.T) {
	_, exitCode, err := Run("bash", "echo ok", "")
	if err != nil {
		t.Fatal(err)
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func TestRunInvalidInterpreter(t *testing.T) {
	_, _, err := Run("nonexistent_interpreter_xyz", "code", "")
	if err == nil {
		t.Error("expected error for invalid interpreter")
	}
}

func TestRunBashWithWorkdir(t *testing.T) {
	dir := t.TempDir()
	output, _, err := Run("bash", "pwd", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, dir) {
		t.Errorf("expected output to contain %q, got %q", dir, output)
	}
}

func TestRunStderrCaptured(t *testing.T) {
	output, _, err := Run("bash", "echo out && echo err >&2", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "out") || !strings.Contains(output, "err") {
		t.Errorf("expected both 'out' and 'err' in output, got %q", output)
	}
}
