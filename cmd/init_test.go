package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	err := Init(file, "My Demo", "v0.3.0")
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.HasPrefix(s, "# My Demo\n\n*") {
		t.Errorf("unexpected content: %q", s)
	}
	if !strings.Contains(s, "T") && !strings.Contains(s, "Z") {
		t.Error("expected ISO 8601 timestamp")
	}
	if !strings.Contains(s, "by Showboat v0.3.0") {
		t.Errorf("expected version in dateline: %q", s)
	}
}

func TestInitErrorsIfExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	os.WriteFile(file, []byte("existing"), 0644)

	err := Init(file, "My Demo", "v0.3.0")
	if err == nil {
		t.Error("expected error when file exists")
	}
}
