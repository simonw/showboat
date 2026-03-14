package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/simonw/showboat/markdown"
)

func TestVerifyPasses(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Exec(file, "bash", "echo hello", ""); err != nil {
		t.Fatal(err)
	}

	diffs, err := Verify(file, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(diffs) != 0 {
		t.Errorf("expected no diffs, got %d: %v", len(diffs), diffs)
	}
}

func TestVerifyDetectsDrift(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Exec(file, "bash", "echo hello", ""); err != nil {
		t.Fatal(err)
	}

	// Tamper with the output block only (not the code block)
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(content), "```output\nhello\n```", "```output\nwrong\n```", 1)
	if err := os.WriteFile(file, []byte(tampered), 0644); err != nil {
		t.Fatal(err)
	}

	diffs, err := Verify(file, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}

	if !strings.Contains(diffs[0].Actual, "hello") {
		t.Errorf("expected actual to contain 'hello', got: %s", diffs[0].Actual)
	}
	if !strings.Contains(diffs[0].Expected, "wrong") {
		t.Errorf("expected expected to contain 'wrong', got: %s", diffs[0].Expected)
	}
}

func TestVerifyWritesOutput(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")
	outputFile := filepath.Join(dir, "updated.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Exec(file, "bash", "echo hello", ""); err != nil {
		t.Fatal(err)
	}

	// Save original content
	original, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with the output block only (not the code block)
	tampered := strings.Replace(string(original), "```output\nhello\n```", "```output\nwrong\n```", 1)
	if err := os.WriteFile(file, []byte(tampered), 0644); err != nil {
		t.Fatal(err)
	}

	diffs, err := Verify(file, outputFile, "")
	if err != nil {
		t.Fatal(err)
	}

	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}

	// Check original is untouched (still tampered)
	currentContent, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(currentContent), "wrong") {
		t.Error("original file should still contain tampered content")
	}

	// Check output file has correct output
	updatedContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updatedContent), "hello") {
		t.Errorf("output file should contain correct output, got: %s", updatedContent)
	}
	if strings.Contains(string(updatedContent), "wrong") {
		t.Errorf("output file should not contain tampered output, got: %s", updatedContent)
	}
}

func TestDiffString(t *testing.T) {
	d := Diff{
		BlockIndex: 3,
		Expected:   "hello\n",
		Actual:     "world\n",
	}

	s := d.String()
	if !strings.Contains(s, "block 3") {
		t.Errorf("expected 'block 3' in Diff.String(), got: %s", s)
	}
	if !strings.Contains(s, "expected: hello") {
		t.Errorf("expected 'expected: hello' in Diff.String(), got: %s", s)
	}
	if !strings.Contains(s, "actual:   world") {
		t.Errorf("expected 'actual:   world' in Diff.String(), got: %s", s)
	}
}

func TestDiffStringTrimsTrailingNewlines(t *testing.T) {
	d := Diff{
		BlockIndex: 0,
		Expected:   "foo\n\n\n",
		Actual:     "bar\n\n",
	}

	s := d.String()
	// Should not end with trailing newlines in expected/actual portions
	if strings.Contains(s, "expected: foo\n\n") {
		t.Errorf("expected trailing newlines to be trimmed, got: %q", s)
	}
}

func TestVerifySkipsImageBlocks(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	// Write a document with an image code block (should be skipped by verify)
	blocks := []markdown.Block{
		markdown.TitleBlock{Title: "Test", Timestamp: "2026-01-01T00:00:00Z", Version: "dev"},
		markdown.CodeBlock{Lang: "bash", Code: "echo hello", IsImage: true},
		markdown.ImageOutputBlock{AltText: "screenshot", Filename: "img.png"},
	}
	if err := writeBlocks(file, blocks); err != nil {
		t.Fatal(err)
	}

	diffs, err := Verify(file, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Errorf("expected no diffs for image-only document, got %d", len(diffs))
	}
}

func TestVerifyNonexistentFile(t *testing.T) {
	_, err := Verify("/nonexistent/path/demo.md", "", "")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestVerifyMultipleBlocks(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Exec(file, "bash", "echo one", ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Exec(file, "bash", "echo two", ""); err != nil {
		t.Fatal(err)
	}

	// Tamper with both outputs
	content, _ := os.ReadFile(file)
	s := string(content)
	s = strings.Replace(s, "```output\none\n```", "```output\nwrong1\n```", 1)
	s = strings.Replace(s, "```output\ntwo\n```", "```output\nwrong2\n```", 1)
	os.WriteFile(file, []byte(s), 0644)

	diffs, err := Verify(file, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(diffs))
	}
}

func TestVerifyCodeBlockWithoutOutput(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	// Write a document with a code block but no following output block
	blocks := []markdown.Block{
		markdown.TitleBlock{Title: "Test", Timestamp: "2026-01-01T00:00:00Z", Version: "dev"},
		markdown.CodeBlock{Lang: "bash", Code: "echo hello"},
	}
	if err := writeBlocks(file, blocks); err != nil {
		t.Fatal(err)
	}

	// Should not error — code block without output is just skipped
	diffs, err := Verify(file, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs for code block without output, got %d", len(diffs))
	}
}
