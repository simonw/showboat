package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPopRemovesCommentary(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}
	if err := Note(file, "Hello world"); err != nil {
		t.Fatal(err)
	}

	if err := Pop(file); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "Hello world") {
		t.Error("expected commentary to be removed after pop")
	}
	if !strings.Contains(string(content), "# Test") {
		t.Error("expected title to remain after pop")
	}
}

func TestPopRemovesCodeAndOutput(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Exec(file, "bash", "echo hello", ""); err != nil {
		t.Fatal(err)
	}

	if err := Pop(file); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if strings.Contains(s, "echo hello") {
		t.Error("expected code block to be removed after pop")
	}
	if strings.Contains(s, "```output") {
		t.Error("expected output block to be removed after pop")
	}
}

func TestPopFailsOnTitleOnly(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	err := Pop(file)
	if err == nil {
		t.Error("expected error when popping title-only document")
	}
	if !strings.Contains(err.Error(), "nothing to pop") {
		t.Errorf("expected 'nothing to pop' error, got: %v", err)
	}
}

func TestPopFailsOnEmptyFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "empty.md")
	if err := os.WriteFile(file, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	err := Pop(file)
	if err == nil {
		t.Error("expected error when popping empty document")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected 'empty' error, got: %v", err)
	}
}

func TestPopNonexistentFile(t *testing.T) {
	err := Pop("/nonexistent/path/demo.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestPopMultipleTimes(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}
	if err := Note(file, "A comment"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Exec(file, "bash", "echo hello", ""); err != nil {
		t.Fatal(err)
	}

	// Pop exec (code + output)
	if err := Pop(file); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(file)
	s := string(content)
	if strings.Contains(s, "echo hello") {
		t.Error("expected code block removed")
	}
	if !strings.Contains(s, "A comment") {
		t.Error("expected commentary to remain")
	}

	// Pop commentary
	if err := Pop(file); err != nil {
		t.Fatal(err)
	}
	content, _ = os.ReadFile(file)
	s = string(content)
	if strings.Contains(s, "A comment") {
		t.Error("expected commentary removed")
	}

	// Pop on title-only should fail
	err := Pop(file)
	if err == nil {
		t.Error("expected error on title-only document")
	}
}
