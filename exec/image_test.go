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

func TestCopyImage(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "photo.png")
	// Write a minimal PNG header so the file exists
	if err := os.WriteFile(imgPath, []byte("\x89PNG\r\n\x1a\n"), 0644); err != nil {
		t.Fatal(err)
	}

	destDir := t.TempDir()
	filename, err := CopyImage(imgPath, destDir)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasSuffix(filename, ".png") {
		t.Errorf("expected .png suffix, got %q", filename)
	}

	destPath := filepath.Join(destDir, filename)
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s", destPath)
	}
}

func TestCopyImageBadPath(t *testing.T) {
	destDir := t.TempDir()
	_, err := CopyImage("/nonexistent/file.png", destDir)
	if err == nil {
		t.Error("expected error for nonexistent image path")
	}
}

func TestCopyImageDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := t.TempDir()
	// Passing a directory as the source should fail
	_, err := CopyImage(tmpDir, destDir)
	if err == nil {
		t.Error("expected error when source is a directory")
	}
	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("expected 'directory' in error message, got: %v", err)
	}
}

func TestCopyImageJpeg(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "photo.jpeg")
	if err := os.WriteFile(imgPath, []byte("fake jpeg"), 0644); err != nil {
		t.Fatal(err)
	}
	destDir := t.TempDir()
	filename, err := CopyImage(imgPath, destDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filename, ".jpeg") {
		t.Errorf("expected .jpeg suffix, got %q", filename)
	}
}

func TestCopyImageSvg(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "icon.svg")
	if err := os.WriteFile(imgPath, []byte("<svg></svg>"), 0644); err != nil {
		t.Fatal(err)
	}
	destDir := t.TempDir()
	filename, err := CopyImage(imgPath, destDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filename, ".svg") {
		t.Errorf("expected .svg suffix, got %q", filename)
	}
}

func TestCopyImageGif(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "anim.gif")
	if err := os.WriteFile(imgPath, []byte("GIF89a"), 0644); err != nil {
		t.Fatal(err)
	}
	destDir := t.TempDir()
	filename, err := CopyImage(imgPath, destDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filename, ".gif") {
		t.Errorf("expected .gif suffix, got %q", filename)
	}
}

func TestRunImageEmptyOutput(t *testing.T) {
	// Script that produces no output
	destDir := t.TempDir()
	_, err := RunImage("true", destDir, "")
	if err == nil {
		t.Error("expected error when script produces no output")
	}
}

func TestRunImageScriptFailure(t *testing.T) {
	destDir := t.TempDir()
	_, err := RunImage("exit 1", destDir, "")
	if err == nil {
		t.Error("expected error when script fails")
	}
}

func TestCopyImageBadExt(t *testing.T) {
	tmpDir := t.TempDir()
	txtPath := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(txtPath, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	destDir := t.TempDir()
	_, err := CopyImage(txtPath, destDir)
	if err == nil {
		t.Error("expected error for unrecognized image format")
	}
}
