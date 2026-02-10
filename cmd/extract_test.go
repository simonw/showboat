package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}
	if err := Note(file, "Hello world"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Exec(file, "bash", "echo hello", ""); err != nil {
		t.Fatal(err)
	}

	commands, err := Extract(file, "")
	if err != nil {
		t.Fatal(err)
	}

	if len(commands) != 3 {
		t.Fatalf("expected 3 commands, got %d: %v", len(commands), commands)
	}

	if !strings.Contains(commands[0], "showboat init") {
		t.Errorf("expected init command, got: %s", commands[0])
	}
	if !strings.Contains(commands[0], file) {
		t.Errorf("expected init command to contain filename %q, got: %s", file, commands[0])
	}
	if !strings.Contains(commands[1], "showboat note") {
		t.Errorf("expected note command, got: %s", commands[1])
	}
	if !strings.Contains(commands[2], "showboat exec") {
		t.Errorf("expected exec command, got: %s", commands[2])
	}
}

func TestExtractOutputOverride(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	commands, err := Extract(file, "other.md")
	if err != nil {
		t.Fatal(err)
	}

	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d: %v", len(commands), commands)
	}

	if !strings.Contains(commands[0], "other.md") {
		t.Errorf("expected output filename in command, got: %s", commands[0])
	}
	if strings.Contains(commands[0], file) {
		t.Errorf("expected original filename to be replaced, got: %s", commands[0])
	}
}

func TestExtractShellQuote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello world", "'hello world'"},
		{"it's", "'it'\\''s'"},
		{"", "''"},
		{"simple", "simple"},
		{"echo $HOME", "'echo $HOME'"},
	}

	for _, tt := range tests {
		got := shellQuote(tt.input)
		if got != tt.expected {
			t.Errorf("shellQuote(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
