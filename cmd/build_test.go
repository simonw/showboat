package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNote(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test"); err != nil {
		t.Fatal(err)
	}

	if err := Note(file, "Hello world"); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "Hello world") {
		t.Errorf("expected commentary in file, got: %s", content)
	}
}

func TestNoteMultiple(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test"); err != nil {
		t.Fatal(err)
	}

	if err := Note(file, "First comment"); err != nil {
		t.Fatal(err)
	}
	if err := Note(file, "Second comment"); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "First comment") {
		t.Errorf("expected first commentary in file, got: %s", s)
	}
	if !strings.Contains(s, "Second comment") {
		t.Errorf("expected second commentary in file, got: %s", s)
	}
}

func TestNoteNoFile(t *testing.T) {
	err := Note("/nonexistent/path/demo.md", "Hello")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestExec(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test"); err != nil {
		t.Fatal(err)
	}

	if _, _, err := Exec(file, "bash", "echo hello", ""); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "```bash\necho hello\n```") {
		t.Errorf("expected code block in file, got: %s", s)
	}
	if !strings.Contains(s, "```output\nhello\n```") {
		t.Errorf("expected output block in file, got: %s", s)
	}
}

func TestExecNonZeroExit(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test"); err != nil {
		t.Fatal(err)
	}

	if _, _, err := Exec(file, "bash", "echo failing && exit 1", ""); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "```bash\necho failing && exit 1\n```") {
		t.Errorf("expected code block in file, got: %s", s)
	}
	if !strings.Contains(s, "```output\nfailing\n```") {
		t.Errorf("expected output block with captured output, got: %s", s)
	}
}

func TestImage(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test"); err != nil {
		t.Fatal(err)
	}

	// Create a tiny valid PNG file and a script that outputs its path
	pngPath := filepath.Join(dir, "test.png")
	// Minimal 1x1 white PNG
	pngData := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG signature
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc,
		0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
	if err := os.WriteFile(pngPath, pngData, 0644); err != nil {
		t.Fatal(err)
	}

	script := "echo " + pngPath

	if err := Image(file, script, ""); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "```bash {image}") {
		t.Errorf("expected image code block in file, got: %s", s)
	}
	if !strings.Contains(s, "![") {
		t.Errorf("expected image output in file, got: %s", s)
	}
}
