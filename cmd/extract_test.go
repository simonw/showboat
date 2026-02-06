package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test"); err != nil {
		t.Fatal(err)
	}
	if err := BuildCommentary(file, "Hello world"); err != nil {
		t.Fatal(err)
	}
	if err := BuildRun(file, "bash", "echo hello", ""); err != nil {
		t.Fatal(err)
	}

	commands, err := Extract(file)
	if err != nil {
		t.Fatal(err)
	}

	if len(commands) != 3 {
		t.Fatalf("expected 3 commands, got %d: %v", len(commands), commands)
	}

	if !strings.Contains(commands[0], "showboat init") {
		t.Errorf("expected init command, got: %s", commands[0])
	}
	if !strings.Contains(commands[1], "commentary") {
		t.Errorf("expected commentary command, got: %s", commands[1])
	}
	if !strings.Contains(commands[2], "run bash") {
		t.Errorf("expected run command, got: %s", commands[2])
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
