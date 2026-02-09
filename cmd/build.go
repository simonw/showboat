package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	execpkg "github.com/simonw/showboat/exec"
	"github.com/simonw/showboat/markdown"
)

// Note appends a commentary block to an existing showboat document.
func Note(file, text string) error {
	blocks, err := readBlocks(file)
	if err != nil {
		return err
	}

	blocks = append(blocks, markdown.CommentaryBlock{Text: text})

	return writeBlocks(file, blocks)
}

// Exec appends a code block, executes it, and appends the output.
// It returns the captured output, the process exit code, and any error.
func Exec(file, lang, code, workdir string) (string, int, error) {
	if _, err := os.Stat(file); err != nil {
		return "", 1, fmt.Errorf("file not found: %s", file)
	}

	output, exitCode, err := execpkg.Run(lang, code, workdir)
	if err != nil {
		return "", exitCode, fmt.Errorf("running code: %w", err)
	}

	blocks, err := readBlocks(file)
	if err != nil {
		return "", exitCode, err
	}

	blocks = append(blocks,
		markdown.CodeBlock{Lang: lang, Code: code},
		markdown.OutputBlock{Content: output},
	)

	if err := writeBlocks(file, blocks); err != nil {
		return output, exitCode, err
	}

	return output, exitCode, nil
}

// Image appends an image code block, runs the script, captures the image.
func Image(file, script, workdir string) error {
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("file not found: %s", file)
	}

	destDir := filepath.Dir(file)
	filename, err := execpkg.RunImage(script, destDir, workdir)
	if err != nil {
		return fmt.Errorf("running image script: %w", err)
	}

	blocks, err := readBlocks(file)
	if err != nil {
		return err
	}

	// Derive alt text from the filename without UUID prefix and date
	altText := strings.TrimSuffix(filename, filepath.Ext(filename))

	blocks = append(blocks,
		markdown.CodeBlock{Lang: "bash", Code: script, IsImage: true},
		markdown.ImageOutputBlock{AltText: altText, Filename: filename},
	)

	return writeBlocks(file, blocks)
}

// readBlocks opens a file and parses its blocks.
func readBlocks(file string) ([]markdown.Block, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	blocks, err := markdown.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	return blocks, nil
}

// writeBlocks creates/truncates a file and writes blocks to it.
func writeBlocks(file string, blocks []markdown.Block) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	return markdown.Write(f, blocks)
}
