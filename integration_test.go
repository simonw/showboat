package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFullWorkflow(t *testing.T) {
	// Build the binary
	tmpBin := filepath.Join(t.TempDir(), "showboat")
	build := exec.Command("go", "build", "-o", tmpBin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	// Init
	run(t, tmpBin, "init", file, "Integration Test Demo")

	// Commentary
	run(t, tmpBin, "build", file, "commentary", "This demo tests the full workflow.")

	// Run bash
	run(t, tmpBin, "build", file, "run", "bash", "echo 'Hello from bash'")

	// Run python
	run(t, tmpBin, "build", file, "run", "python3", "print(2 + 2)")

	// More commentary
	run(t, tmpBin, "build", file, "commentary", "Everything works.")

	// Read the file and check structure
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	checks := []string{
		"# Integration Test Demo",
		"This demo tests the full workflow.",
		"echo 'Hello from bash'",
		"Hello from bash",
		"print(2 + 2)",
		"4",
		"Everything works.",
	}
	for _, check := range checks {
		if !strings.Contains(s, check) {
			t.Errorf("missing expected content: %q\n\nFull file:\n%s", check, s)
		}
	}

	// Verify should pass
	run(t, tmpBin, "verify", file)

	// Extract should produce commands
	out := runOutput(t, tmpBin, "extract", file)
	if !strings.Contains(out, "showboat init") {
		t.Errorf("extract missing init: %s", out)
	}
	if !strings.Contains(out, "run bash") {
		t.Errorf("extract missing run bash: %s", out)
	}

	// Test stdin for commentary
	stdinFile := filepath.Join(dir, "stdin-demo.md")
	run(t, tmpBin, "init", stdinFile, "Stdin Test")

	stdinCmd := exec.Command(tmpBin, "build", stdinFile, "commentary")
	stdinCmd.Stdin = strings.NewReader("Commentary from stdin")
	if out, err := stdinCmd.CombinedOutput(); err != nil {
		t.Fatalf("stdin commentary failed: %s\n%s", err, out)
	}
	stdinContent, _ := os.ReadFile(stdinFile)
	if !strings.Contains(string(stdinContent), "Commentary from stdin") {
		t.Error("stdin commentary not found in file")
	}

	// Tamper and verify should fail
	tampered := strings.Replace(s, "Hello from bash\n", "TAMPERED\n", 1)
	os.WriteFile(file, []byte(tampered), 0644)

	cmd := exec.Command(tmpBin, "verify", file)
	if err := cmd.Run(); err == nil {
		t.Error("expected verify to fail after tampering")
	}

	// Verify with --output should produce corrected file
	outputFile := filepath.Join(dir, "fixed.md")
	cmd = exec.Command(tmpBin, "verify", file, "--output", outputFile)
	cmd.Run() // may exit non-zero, that's fine

	fixed, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(fixed), "Hello from bash") {
		t.Error("expected fixed file to have correct output")
	}
}

func run(t *testing.T, bin string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("command failed: %s %v\n%s\n%s", bin, args, err, out)
	}
}

func runOutput(t *testing.T, bin string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %s %v\n%s\n%s", bin, args, err, out)
	}
	return string(out)
}
