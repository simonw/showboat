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

func TestCopyImageWebP(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "photo.webp")
	// Valid 1x1 pixel WebP image (62 bytes, VP8 encoding)
	webpData := []byte{
		0x52, 0x49, 0x46, 0x46, 0x36, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50, 0x56, 0x50, 0x38, 0x20,
		0x2a, 0x00, 0x00, 0x00, 0xf0, 0x01, 0x00, 0x9d, 0x01, 0x2a, 0x01, 0x00, 0x01, 0x00, 0x02, 0x00,
		0x34, 0x25, 0xa0, 0x02, 0x74, 0xba, 0x01, 0xf8, 0x00, 0x05, 0xf4, 0x00, 0x00, 0x9b, 0xff, 0xcb,
		0x3d, 0xe6, 0x37, 0x7b, 0xa6, 0xff, 0xe2, 0xce, 0xee, 0x96, 0x75, 0xd0, 0x00, 0x00,
	}
	if err := os.WriteFile(imgPath, webpData, 0644); err != nil {
		t.Fatal(err)
	}

	destDir := t.TempDir()
	filename, err := CopyImage(imgPath, destDir)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasSuffix(filename, ".webp") {
		t.Errorf("expected .webp suffix, got %q", filename)
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
