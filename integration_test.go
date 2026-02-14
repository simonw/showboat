package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
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

	// Note
	run(t, tmpBin, "note", file, "This demo tests the full workflow.")

	// Exec bash
	run(t, tmpBin, "exec", file, "bash", "echo 'Hello from bash'")

	// Exec python
	run(t, tmpBin, "exec", file, "python3", "print(2 + 2)")

	// More commentary
	run(t, tmpBin, "note", file, "Everything works.")

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
	if !strings.Contains(out, "exec") {
		t.Errorf("extract missing exec: %s", out)
	}

	// Test stdin for commentary
	stdinFile := filepath.Join(dir, "stdin-demo.md")
	run(t, tmpBin, "init", stdinFile, "Stdin Test")

	stdinCmd := exec.Command(tmpBin, "note", stdinFile)
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

func TestBuildRunOutputAndExitCode(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "showboat")
	build := exec.Command("go", "build", "-o", tmpBin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	run(t, tmpBin, "init", file, "Output Test")

	// Successful command: should print output and exit 0
	out := runOutput(t, tmpBin, "exec", file, "bash", "echo hello world")
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected build run to print output, got: %q", out)
	}

	// Failing command: should print output and exit non-zero
	cmd := exec.Command(tmpBin, "exec", file, "bash", "echo fail output && exit 42")
	failOut, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("expected build run to exit non-zero for failing command")
	}
	if !strings.Contains(string(failOut), "fail output") {
		t.Errorf("expected failing build run to print output, got: %q", string(failOut))
	}
	// Check the exit code is 42
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 42 {
			t.Errorf("expected exit code 42, got %d", exitErr.ExitCode())
		}
	}

	// The failing output should still be in the document
	content, _ := os.ReadFile(file)
	if !strings.Contains(string(content), "fail output") {
		t.Error("expected failing command output to be captured in document")
	}
}

func TestPop(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "showboat")
	build := exec.Command("go", "build", "-o", tmpBin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	// Init and add some content
	run(t, tmpBin, "init", file, "Pop Test")
	run(t, tmpBin, "note", file, "First comment.")
	run(t, tmpBin, "exec", file, "bash", "echo hello")
	run(t, tmpBin, "note", file, "Second comment.")

	// Pop should remove the last commentary
	run(t, tmpBin, "pop", file)
	content, _ := os.ReadFile(file)
	s := string(content)
	if strings.Contains(s, "Second comment.") {
		t.Error("expected pop to remove last commentary")
	}
	if !strings.Contains(s, "hello") {
		t.Error("expected earlier content to remain after pop")
	}

	// Pop again should remove the run entry (code + output)
	run(t, tmpBin, "pop", file)
	content, _ = os.ReadFile(file)
	s = string(content)
	if strings.Contains(s, "echo hello") {
		t.Error("expected pop to remove code block")
	}
	if strings.Contains(s, "hello\n") {
		// Check that the output block was also removed, but "hello" might still
		// appear in the test title check — look more specifically
		if strings.Contains(s, "```output") {
			t.Error("expected pop to remove output block")
		}
	}
	if !strings.Contains(s, "First comment.") {
		t.Error("expected first commentary to remain after popping run")
	}

	// Pop again should remove the first commentary
	run(t, tmpBin, "pop", file)
	content, _ = os.ReadFile(file)
	s = string(content)
	if strings.Contains(s, "First comment.") {
		t.Error("expected pop to remove first commentary")
	}

	// Pop on title-only document should fail
	cmd := exec.Command(tmpBin, "pop", file)
	if err := cmd.Run(); err == nil {
		t.Error("expected pop to fail on title-only document")
	}
}

func TestServerCommand(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "showboat")
	build := exec.Command("go", "build", "-o", tmpBin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	// Create a document with a server block
	run(t, tmpBin, "init", file, "Server Test")
	run(t, tmpBin, "note", file, "Starting a server.")

	// Manually write a server block into the file
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	serverDoc := string(content) + "\n```bash {server}\npython3 -m http.server $PORT\n```\n"
	if err := os.WriteFile(file, []byte(serverDoc), 0644); err != nil {
		t.Fatal(err)
	}

	// Start the server command in the background and read its stdout
	serverCmd := exec.Command(tmpBin, "server", file)
	stdout, err := serverCmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := serverCmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Read the "Server running on port ..." line
	buf := make([]byte, 256)
	n, err := stdout.Read(buf)
	if err != nil {
		serverCmd.Process.Kill()
		serverCmd.Wait()
		t.Fatalf("reading server output: %v", err)
	}
	serverOutput := string(buf[:n])
	if !strings.Contains(serverOutput, "Server running on port") {
		serverCmd.Process.Kill()
		serverCmd.Wait()
		t.Fatalf("expected 'Server running on port' output, got: %q", serverOutput)
	}

	// Kill the server process tree
	serverCmd.Process.Signal(syscall.SIGTERM)
	serverCmd.Wait()

	// Test that extract outputs a server command for the server block
	extractOut := runOutput(t, tmpBin, "extract", file)
	if !strings.Contains(extractOut, "showboat server") {
		t.Errorf("extract should include 'showboat server', got: %s", extractOut)
	}

	// Test that verify works with server blocks
	run(t, tmpBin, "verify", file)
}

func TestVerifyWithServerIntegration(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "showboat")
	build := exec.Command("go", "build", "-o", tmpBin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	// Create a document with a server block and a curl that hits it
	doc := `# Server Verify Test

*2026-02-14T00:00:00Z*

` + "```bash {server}\npython3 -m http.server $PORT\n```" + `

` + "```bash\ncurl -s -o /dev/null -w '%{http_code}\\n' http://localhost:$PORT/\n```" + `

` + "```output\n200\n```\n"

	if err := os.WriteFile(file, []byte(doc), 0644); err != nil {
		t.Fatal(err)
	}

	// Verify should pass (server starts, curl hits it, output matches)
	run(t, tmpBin, "verify", file)
}

func TestVersionFlagDefault(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "showcase")
	build := exec.Command("go", "build", "-o", tmpBin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	got := strings.TrimSpace(runOutput(t, tmpBin, "--version"))
	if got != "dev" {
		t.Fatalf("expected version dev, got %q", got)
	}
}

func TestVersionFlagInjectedByLdflags(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "showcase")
	build := exec.Command("go", "build", "-ldflags", "-X main.version=1.2.3", "-o", tmpBin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	got := strings.TrimSpace(runOutput(t, tmpBin, "--version"))
	if got != "1.2.3" {
		t.Fatalf("expected version 1.2.3, got %q", got)
	}
}
