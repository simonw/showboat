package exec

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// validImageExts lists recognized image file extensions.
var validImageExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".webp": true,
}

// CopyImage copies an image file to destDir with a generated
// <uuid>-<date>.<ext> filename. It validates that srcPath exists, is a
// regular file, and has a recognized image extension.
// Returns the new filename (not the full path).
func CopyImage(srcPath, destDir string) (string, error) {
	// Verify file exists
	info, err := os.Stat(srcPath)
	if err != nil {
		return "", fmt.Errorf("image file not found: %s", srcPath)
	}
	if info.IsDir() {
		return "", fmt.Errorf("image path is a directory: %s", srcPath)
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(srcPath))
	if !validImageExts[ext] {
		return "", fmt.Errorf("unrecognized image format: %s", ext)
	}

	// Generate destination filename
	id := uuid.New().String()[:8]
	date := time.Now().UTC().Format("2006-01-02")
	newFilename := fmt.Sprintf("%s-%s%s", id, date, ext)

	// Copy file
	src, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("opening image: %w", err)
	}
	defer src.Close()

	dstPath := filepath.Join(destDir, newFilename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("creating destination: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("copying image: %w", err)
	}

	return newFilename, nil
}

// RunImage runs a bash script that is expected to produce an image file.
// The last line of stdout is treated as the path to the image.
// The image is copied to destDir with a <uuid>-<date>.<ext> filename.
// Returns the new filename (not the full path).
func RunImage(script, destDir, workdir string) (string, error) {
	output, _, err := Run("bash", script, workdir)
	if err != nil {
		return "", fmt.Errorf("running image script: %w", err)
	}

	// Last non-empty line of output is the image path
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("image script produced no output")
	}
	srcPath := strings.TrimSpace(lines[len(lines)-1])

	return CopyImage(srcPath, destDir)
}
