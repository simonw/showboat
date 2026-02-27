package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyPasses(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Exec(ExecOpts{File: file, Lang: "bash", Code: "echo hello"}); err != nil {
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
	if _, _, err := Exec(ExecOpts{File: file, Lang: "bash", Code: "echo hello"}); err != nil {
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

func TestVerifyWithFilter(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Exec(ExecOpts{File: file, Lang: "python", Code: "hello\n", Filter: "cat"}); err != nil {
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

func TestVerifyWritesOutput(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")
	outputFile := filepath.Join(dir, "updated.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Exec(ExecOpts{File: file, Lang: "bash", Code: "echo hello"}); err != nil {
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
