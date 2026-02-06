package exec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunImageScript(t *testing.T) {
	// Create a temp dir for the test
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.png")

	// Script that creates a tiny valid PNG and prints its path
	script := `printf '\x89PNG\r\n\x1a\n' > ` + imgPath + ` && echo ` + imgPath

	destDir := t.TempDir()
	filename, err := RunImage(script, destDir, "")
	if err != nil {
		t.Fatal(err)
	}

	// Filename should match <uuid>-<date>.<ext> pattern
	if !strings.HasSuffix(filename, ".png") {
		t.Errorf("expected .png suffix, got %q", filename)
	}

	// File should exist in destDir
	destPath := filepath.Join(destDir, filename)
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s", destPath)
	}
}

func TestRunImageScriptBadPath(t *testing.T) {
	script := `echo /nonexistent/file.png`
	destDir := t.TempDir()
	_, err := RunImage(script, destDir, "")
	if err == nil {
		t.Error("expected error for nonexistent image path")
	}
}
