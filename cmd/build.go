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

	newBlock := markdown.CommentaryBlock{Text: text}
	blocks = append(blocks, newBlock)

	if err := writeBlocks(file, blocks); err != nil {
		return err
	}

	docID := documentID(blocks)
	if docID != "" {
		postSection(docID, "note", []markdown.Block{newBlock})
	}
	return nil
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

	codeBlock := markdown.CodeBlock{Lang: lang, Code: code}
	outputBlock := markdown.OutputBlock{Content: output}
	blocks = append(blocks, codeBlock, outputBlock)

	if err := writeBlocks(file, blocks); err != nil {
		return output, exitCode, err
	}

	docID := documentID(blocks)
	if docID != "" {
		postSection(docID, "exec", []markdown.Block{codeBlock, outputBlock})
	}

	return output, exitCode, nil
}

// Image appends an image reference to a showboat document. The input is either
// a plain path to an image file or a markdown image reference of the form
// ![alt text](path). When a markdown reference is provided the alt text is
// preserved; otherwise it is derived from the generated filename.
func Image(file, input, workdir string) error {
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("file not found: %s", file)
	}

	imgPath, altText := parseImageInput(input)

	destDir := filepath.Dir(file)
	filename, err := execpkg.CopyImage(imgPath, destDir)
	if err != nil {
		return err
	}

	blocks, err := readBlocks(file)
	if err != nil {
		return err
	}

	if altText == "" {
		// Derive alt text from the filename without UUID prefix and date
		altText = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	codeBlock := markdown.CodeBlock{Lang: "bash", Code: input, IsImage: true}
	imgBlock := markdown.ImageOutputBlock{AltText: altText, Filename: filename}
	blocks = append(blocks, codeBlock, imgBlock)

	if err := writeBlocks(file, blocks); err != nil {
		return err
	}

	docID := documentID(blocks)
	if docID != "" {
		copiedImagePath := filepath.Join(destDir, filename)
		postImage(docID, []markdown.Block{codeBlock, imgBlock}, copiedImagePath)
	}
	return nil
}

// parseImageInput checks whether input is a markdown image reference
// (![alt](path)) or a plain file path. It returns the image path and any
// extracted alt text (empty when the input is a plain path).
// It also handles the common case where the shell escapes "!" to "\!".
func parseImageInput(input string) (path, altText string) {
	trimmed := strings.TrimSpace(input)
	// Some shells escape "!" to "\!", so strip the leading backslash.
	if strings.HasPrefix(trimmed, `\![`) {
		trimmed = trimmed[1:]
	}
	if strings.HasPrefix(trimmed, "![") && strings.HasSuffix(trimmed, ")") {
		// Extract alt text between ![ and ]
		rest := trimmed[2:]
		closeBracket := strings.Index(rest, "](")
		if closeBracket != -1 {
			altText = rest[:closeBracket]
			path = rest[closeBracket+2 : len(rest)-1]
			return path, altText
		}
	}
	return trimmed, ""
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
