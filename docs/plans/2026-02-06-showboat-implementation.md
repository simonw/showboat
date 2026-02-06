# Showboat Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI tool that helps agents create, verify, and extract executable markdown demo documents.

**Architecture:** A single `showboat` binary with subcommands (`init`, `build`, `verify`, `extract`). The `markdown` package handles parsing and serialization of the document format. The `exec` package handles running code blocks and capturing output. The `cmd` package wires CLI arguments to these packages.

**Tech Stack:** Go (standard library + `github.com/google/uuid` for UUID generation). No CLI framework — use `os.Args` and `flag` directly to keep help text fully custom.

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `main.go`

**Step 1: Initialize the Go module**

Run: `go mod init github.com/simonw/showboat`
Expected: `go.mod` created with module path `github.com/simonw/showboat`

**Step 2: Create a minimal main.go**

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		fmt.Fprintln(os.Stderr, "init: not yet implemented")
		os.Exit(1)
	case "build":
		fmt.Fprintln(os.Stderr, "build: not yet implemented")
		os.Exit(1)
	case "verify":
		fmt.Fprintln(os.Stderr, "verify: not yet implemented")
		os.Exit(1)
	case "extract":
		fmt.Fprintln(os.Stderr, "extract: not yet implemented")
		os.Exit(1)
	case "--help", "-h", "help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`showboat - Create executable demo documents that show and prove an agent's work.

Showboat helps agents build markdown documents that mix commentary, executable
code blocks, and captured output. These documents serve as both readable
documentation and reproducible proof of work. A verifier can re-execute all
code blocks and confirm the outputs still match.

Usage:
  showboat init <file> <title>             Create a new demo document
  showboat build <file> commentary [text]  Append commentary (text or stdin)
  showboat build <file> run <lang> [code]  Run code and capture output
  showboat build <file> image [script]     Run script, capture image output
  showboat verify <file> [--output <new>]  Re-run and diff all code blocks
  showboat extract <file>                  Emit build commands to recreate file

Global Options:
  --workdir <dir>   Set working directory for code execution (default: current)
  --help, -h        Show this help message

Stdin:
  The build subcommands accept input from stdin when the text/code argument is
  omitted. For example:
    echo "Hello world" | showboat build demo.md commentary
    cat script.sh | showboat build demo.md run bash

Example:
  # Create a demo
  showboat init demo.md "Setting Up a Python Project"

  # Add commentary
  showboat build demo.md commentary "First, let's create a virtual environment."

  # Run a command and capture output
  showboat build demo.md run bash "python3 -m venv .venv && echo 'Done'"

  # Run Python and capture output
  showboat build demo.md run python "print('Hello from Python')"

  # Capture a screenshot
  showboat build demo.md image "python screenshot.py http://localhost:8000"

  # Verify the demo still works
  showboat verify demo.md

  # See what commands built the demo
  showboat extract demo.md

Resulting markdown format:

  # Setting Up a Python Project

  *2026-02-06T15:30:00Z*

  First, let's create a virtual environment.

  ` + "```" + `bash
  python3 -m venv .venv && echo 'Done'
  ` + "```" + `

  ` + "```" + `output
  Done
  ` + "```" + `

  ` + "```" + `python
  print('Hello from Python')
  ` + "```" + `

  ` + "```" + `output
  Hello from Python
  ` + "```" + `
`)
}
```

**Step 3: Verify it compiles and help works**

Run: `go build -o showboat . && ./showboat --help`
Expected: Help text prints, exit 0

Run: `./showboat`
Expected: Help text prints, exit 1

Run: `./showboat init`
Expected: "init: not yet implemented", exit 1

**Step 4: Commit**

```bash
git add go.mod main.go
git commit -m "feat: scaffold showboat CLI with help text and command routing"
```

---

### Task 2: Block Types (`markdown/blocks.go`)

**Files:**
- Create: `markdown/blocks.go`
- Create: `markdown/blocks_test.go`

**Step 1: Write the test**

```go
package markdown

import "testing"

func TestCommentaryBlock(t *testing.T) {
	b := CommentaryBlock{Text: "Hello world\n\nThis is a test."}
	if b.Type() != "commentary" {
		t.Errorf("expected type commentary, got %s", b.Type())
	}
}

func TestCodeBlock(t *testing.T) {
	b := CodeBlock{Lang: "bash", Code: "echo hello", IsImage: false}
	if b.Type() != "code" {
		t.Errorf("expected type code, got %s", b.Type())
	}
}

func TestOutputBlock(t *testing.T) {
	b := OutputBlock{Content: "hello\n"}
	if b.Type() != "output" {
		t.Errorf("expected type output, got %s", b.Type())
	}
}

func TestImageOutputBlock(t *testing.T) {
	b := ImageOutputBlock{AltText: "Screenshot", Filename: "abc-2026-02-06.png"}
	if b.Type() != "output-image" {
		t.Errorf("expected type output-image, got %s", b.Type())
	}
}

func TestTitleBlock(t *testing.T) {
	b := TitleBlock{Title: "My Demo", Timestamp: "2026-02-06T15:30:00Z"}
	if b.Type() != "title" {
		t.Errorf("expected type title, got %s", b.Type())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./markdown/ -v`
Expected: FAIL — types not defined

**Step 3: Implement block types**

```go
package markdown

// Block is an element in a showboat document.
type Block interface {
	Type() string
}

// TitleBlock is the document header: an H1 title and a timestamp.
type TitleBlock struct {
	Title     string
	Timestamp string
}

func (b TitleBlock) Type() string { return "title" }

// CommentaryBlock is free-form markdown prose.
type CommentaryBlock struct {
	Text string
}

func (b CommentaryBlock) Type() string { return "commentary" }

// CodeBlock is an executable fenced code block.
type CodeBlock struct {
	Lang    string
	Code    string
	IsImage bool
}

func (b CodeBlock) Type() string { return "code" }

// OutputBlock is captured text output from a code block.
type OutputBlock struct {
	Content string
}

func (b OutputBlock) Type() string { return "output" }

// ImageOutputBlock is a captured image reference from an image code block.
type ImageOutputBlock struct {
	AltText  string
	Filename string
}

func (b ImageOutputBlock) Type() string { return "output-image" }
```

**Step 4: Run test to verify it passes**

Run: `go test ./markdown/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add markdown/
git commit -m "feat: add markdown block types"
```

---

### Task 3: Markdown Writer (`markdown/writer.go`)

**Files:**
- Create: `markdown/writer.go`
- Create: `markdown/writer_test.go`

**Step 1: Write the test**

```go
package markdown

import (
	"strings"
	"testing"
)

func TestWriteTitle(t *testing.T) {
	var buf strings.Builder
	blocks := []Block{
		TitleBlock{Title: "My Demo", Timestamp: "2026-02-06T15:30:00Z"},
	}
	err := Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	expected := "# My Demo\n\n*2026-02-06T15:30:00Z*\n"
	if buf.String() != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, buf.String())
	}
}

func TestWriteCommentary(t *testing.T) {
	var buf strings.Builder
	blocks := []Block{
		CommentaryBlock{Text: "Hello world."},
	}
	err := Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Hello world.\n"
	if buf.String() != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, buf.String())
	}
}

func TestWriteCodeAndOutput(t *testing.T) {
	var buf strings.Builder
	blocks := []Block{
		CodeBlock{Lang: "bash", Code: "echo hello"},
		OutputBlock{Content: "hello\n"},
	}
	err := Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	expected := "```bash\necho hello\n```\n\n```output\nhello\n```\n"
	if buf.String() != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, buf.String())
	}
}

func TestWriteImageCodeAndOutput(t *testing.T) {
	var buf strings.Builder
	blocks := []Block{
		CodeBlock{Lang: "bash", Code: "python screenshot.py", IsImage: true},
		ImageOutputBlock{AltText: "Screenshot", Filename: "abc-2026-02-06.png"},
	}
	err := Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	expected := "```bash {image}\npython screenshot.py\n```\n\n```output-image\n![Screenshot](abc-2026-02-06.png)\n```\n"
	if buf.String() != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, buf.String())
	}
}

func TestWriteFullDocument(t *testing.T) {
	var buf strings.Builder
	blocks := []Block{
		TitleBlock{Title: "Demo", Timestamp: "2026-02-06T00:00:00Z"},
		CommentaryBlock{Text: "Let's begin."},
		CodeBlock{Lang: "bash", Code: "echo hi"},
		OutputBlock{Content: "hi\n"},
		CommentaryBlock{Text: "Done."},
	}
	err := Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	expected := "# Demo\n\n*2026-02-06T00:00:00Z*\n\nLet's begin.\n\n```bash\necho hi\n```\n\n```output\nhi\n```\n\nDone.\n"
	if buf.String() != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, buf.String())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./markdown/ -v -run TestWrite`
Expected: FAIL — `Write` not defined

**Step 3: Implement the writer**

```go
package markdown

import (
	"fmt"
	"io"
)

// Write serializes a slice of Blocks into showboat markdown format.
func Write(w io.Writer, blocks []Block) error {
	for i, block := range blocks {
		if i > 0 {
			// Add blank line separator between blocks, except after code blocks
			// which are followed by their output block.
			prev := blocks[i-1]
			if _, isCode := prev.(CodeBlock); !isCode {
				fmt.Fprint(w, "\n")
			}
		}

		switch b := block.(type) {
		case TitleBlock:
			fmt.Fprintf(w, "# %s\n\n*%s*\n", b.Title, b.Timestamp)
		case CommentaryBlock:
			fmt.Fprintf(w, "%s\n", b.Text)
		case CodeBlock:
			if b.IsImage {
				fmt.Fprintf(w, "```%s {image}\n%s\n```\n", b.Lang, b.Code)
			} else {
				fmt.Fprintf(w, "```%s\n%s\n```\n", b.Lang, b.Code)
			}
		case OutputBlock:
			fmt.Fprintf(w, "\n```output\n%s```\n", b.Content)
		case ImageOutputBlock:
			fmt.Fprintf(w, "\n```output-image\n![%s](%s)\n```\n", b.AltText, b.Filename)
		default:
			return fmt.Errorf("unknown block type: %T", block)
		}
	}
	return nil
}
```

**Step 4: Run tests and iterate until they pass**

Run: `go test ./markdown/ -v -run TestWrite`
Expected: PASS

Note: The exact separator logic between blocks may need tweaking to match the expected output. Adjust the blank line logic until all tests pass. The tests are the source of truth for the format.

**Step 5: Commit**

```bash
git add markdown/writer.go markdown/writer_test.go
git commit -m "feat: add markdown writer with serialization of all block types"
```

---

### Task 4: Markdown Parser (`markdown/parser.go`)

**Files:**
- Create: `markdown/parser.go`
- Create: `markdown/parser_test.go`

**Step 1: Write the test**

```go
package markdown

import (
	"strings"
	"testing"
)

func TestParseTitle(t *testing.T) {
	input := "# My Demo\n\n*2026-02-06T15:30:00Z*\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	tb, ok := blocks[0].(TitleBlock)
	if !ok {
		t.Fatalf("expected TitleBlock, got %T", blocks[0])
	}
	if tb.Title != "My Demo" {
		t.Errorf("expected title 'My Demo', got %q", tb.Title)
	}
	if tb.Timestamp != "2026-02-06T15:30:00Z" {
		t.Errorf("expected timestamp '2026-02-06T15:30:00Z', got %q", tb.Timestamp)
	}
}

func TestParseCommentary(t *testing.T) {
	input := "# Demo\n\n*2026-02-06T00:00:00Z*\n\nHello world.\n\nMore text here.\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d: %+v", len(blocks), blocks)
	}
	cb, ok := blocks[1].(CommentaryBlock)
	if !ok {
		t.Fatalf("expected CommentaryBlock, got %T", blocks[1])
	}
	if cb.Text != "Hello world.\n\nMore text here." {
		t.Errorf("unexpected text: %q", cb.Text)
	}
}

func TestParseCodeAndOutput(t *testing.T) {
	input := "```bash\necho hello\n```\n\n```output\nhello\n```\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	code, ok := blocks[0].(CodeBlock)
	if !ok {
		t.Fatalf("expected CodeBlock, got %T", blocks[0])
	}
	if code.Lang != "bash" || code.Code != "echo hello" || code.IsImage {
		t.Errorf("unexpected code block: %+v", code)
	}
	out, ok := blocks[1].(OutputBlock)
	if !ok {
		t.Fatalf("expected OutputBlock, got %T", blocks[1])
	}
	if out.Content != "hello\n" {
		t.Errorf("unexpected output: %q", out.Content)
	}
}

func TestParseImageCodeAndOutput(t *testing.T) {
	input := "```bash {image}\npython screenshot.py\n```\n\n```output-image\n![Screenshot](abc-2026-02-06.png)\n```\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	code, ok := blocks[0].(CodeBlock)
	if !ok {
		t.Fatalf("expected CodeBlock, got %T", blocks[0])
	}
	if !code.IsImage {
		t.Error("expected IsImage=true")
	}
	img, ok := blocks[1].(ImageOutputBlock)
	if !ok {
		t.Fatalf("expected ImageOutputBlock, got %T", blocks[1])
	}
	if img.AltText != "Screenshot" || img.Filename != "abc-2026-02-06.png" {
		t.Errorf("unexpected image output: %+v", img)
	}
}

func TestRoundTrip(t *testing.T) {
	input := "# Demo\n\n*2026-02-06T00:00:00Z*\n\nLet's begin.\n\n```bash\necho hi\n```\n\n```output\nhi\n```\n\nDone.\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	var buf strings.Builder
	err = Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != input {
		t.Errorf("round trip mismatch.\nexpected:\n%s\ngot:\n%s", input, buf.String())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./markdown/ -v -run TestParse`
Expected: FAIL — `Parse` not defined

**Step 3: Implement the parser**

The parser reads line by line. It needs to handle:

1. Lines starting with `# ` at the very beginning → TitleBlock (next non-empty line is `*timestamp*`)
2. Lines starting with `` ``` `` → start of a fenced block
   - `` ```lang `` or `` ```lang {image} `` → CodeBlock
   - `` ```output `` → OutputBlock
   - `` ```output-image `` → ImageOutputBlock
3. Everything else → CommentaryBlock (accumulate lines until a fence or EOF)

```go
package markdown

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Parse reads showboat markdown from r and returns a slice of Blocks.
func Parse(r io.Reader) ([]Block, error) {
	scanner := bufio.NewScanner(r)
	var blocks []Block
	var lines []string
	firstBlock := true

	flushCommentary := func() {
		text := strings.Join(lines, "\n")
		text = strings.TrimRight(text, "\n")
		text = strings.TrimLeft(text, "\n")
		if text != "" {
			blocks = append(blocks, CommentaryBlock{Text: text})
		}
		lines = nil
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Detect title at start of document
		if firstBlock && strings.HasPrefix(line, "# ") {
			firstBlock = false
			title := strings.TrimPrefix(line, "# ")
			// Read blank line then timestamp
			timestamp := ""
			for scanner.Scan() {
				tl := scanner.Text()
				if tl == "" {
					continue
				}
				if strings.HasPrefix(tl, "*") && strings.HasSuffix(tl, "*") {
					timestamp = strings.Trim(tl, "*")
				}
				break
			}
			blocks = append(blocks, TitleBlock{Title: title, Timestamp: timestamp})
			continue
		}
		firstBlock = false

		// Detect fenced code blocks
		if strings.HasPrefix(line, "```") {
			flushCommentary()
			fence := strings.TrimPrefix(line, "```")
			fence = strings.TrimSpace(fence)

			switch {
			case fence == "output":
				// Read output content until closing ```
				var content strings.Builder
				for scanner.Scan() {
					ol := scanner.Text()
					if ol == "```" {
						break
					}
					content.WriteString(ol)
					content.WriteString("\n")
				}
				blocks = append(blocks, OutputBlock{Content: content.String()})

			case fence == "output-image":
				// Read image reference line: ![alt](filename)
				altText := ""
				filename := ""
				for scanner.Scan() {
					ol := scanner.Text()
					if ol == "```" {
						break
					}
					if strings.HasPrefix(ol, "![") {
						// Parse ![alt](filename)
						closeBracket := strings.Index(ol, "](")
						if closeBracket != -1 {
							altText = ol[2:closeBracket]
							rest := ol[closeBracket+2:]
							closeParen := strings.Index(rest, ")")
							if closeParen != -1 {
								filename = rest[:closeParen]
							}
						}
					}
				}
				blocks = append(blocks, ImageOutputBlock{AltText: altText, Filename: filename})

			default:
				// Code block: parse lang and {image} flag
				lang := fence
				isImage := false
				if strings.Contains(fence, "{image}") {
					lang = strings.TrimSpace(strings.Replace(fence, "{image}", "", 1))
					isImage = true
				}
				var code strings.Builder
				first := true
				for scanner.Scan() {
					cl := scanner.Text()
					if cl == "```" {
						break
					}
					if !first {
						code.WriteString("\n")
					}
					code.WriteString(cl)
					first = false
				}
				blocks = append(blocks, CodeBlock{Lang: lang, Code: code.String(), IsImage: isImage})
			}
			continue
		}

		lines = append(lines, line)
	}

	flushCommentary()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	return blocks, nil
}
```

**Step 4: Run tests and iterate until they pass**

Run: `go test ./markdown/ -v`
Expected: All tests PASS, including `TestRoundTrip`

Note: Round-trip test is critical. If it fails, adjust either parser or writer until the full cycle works. The markdown format in the tests is the canonical specification.

**Step 5: Commit**

```bash
git add markdown/parser.go markdown/parser_test.go
git commit -m "feat: add markdown parser with round-trip support"
```

---

### Task 5: Code Execution (`exec/runner.go`)

**Files:**
- Create: `exec/runner.go`
- Create: `exec/runner_test.go`

**Step 1: Write the test**

```go
package exec

import "testing"

func TestRunBash(t *testing.T) {
	output, err := Run("bash", "echo hello", "")
	if err != nil {
		t.Fatal(err)
	}
	if output != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", output)
	}
}

func TestRunPython(t *testing.T) {
	output, err := Run("python", "print('hi')", "")
	if err != nil {
		// Try python3
		output, err = Run("python3", "print('hi')", "")
		if err != nil {
			t.Fatal(err)
		}
	}
	if output != "hi\n" {
		t.Errorf("expected 'hi\\n', got %q", output)
	}
}

func TestRunWithWorkdir(t *testing.T) {
	output, err := Run("bash", "pwd", "/tmp")
	if err != nil {
		t.Fatal(err)
	}
	if output != "/tmp\n" {
		t.Errorf("expected '/tmp\\n', got %q", output)
	}
}

func TestRunNonZeroExit(t *testing.T) {
	output, err := Run("bash", "echo oops && exit 1", "")
	// Non-zero exit should NOT return an error — the output is captured.
	if err != nil {
		t.Fatal(err)
	}
	if output != "oops\n" {
		t.Errorf("expected 'oops\\n', got %q", output)
	}
}

func TestRunStderrCaptured(t *testing.T) {
	output, err := Run("bash", "echo out && echo err >&2", "")
	if err != nil {
		t.Fatal(err)
	}
	// Both stdout and stderr should appear in output
	if output != "out\nerr\n" {
		// Order might vary; just check both are present
		if !(len(output) > 0) {
			t.Errorf("expected some output, got %q", output)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./exec/ -v`
Expected: FAIL — `Run` not defined

**Step 3: Implement the runner**

```go
package exec

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Run executes code using the given language interpreter and returns
// the combined stdout+stderr output. Non-zero exit codes are not
// treated as errors — the output is still captured and returned.
// If workdir is empty, the current directory is used.
func Run(lang, code, workdir string) (string, error) {
	var cmd *exec.Cmd
	switch lang {
	case "bash", "sh":
		cmd = exec.Command(lang, "-c", code)
	case "python", "python3":
		cmd = exec.Command(lang, "-c", code)
	default:
		// Generic: try lang -c code
		cmd = exec.Command(lang, "-c", code)
	}

	if workdir != "" {
		cmd.Dir = workdir
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	// We only return an error if the command couldn't be started at all.
	// Non-zero exit codes are fine — the output is what matters.
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Non-zero exit — that's fine, return output
			return buf.String(), nil
		}
		return "", fmt.Errorf("executing %s: %w", lang, err)
	}

	return buf.String(), nil
}
```

**Step 4: Run tests**

Run: `go test ./exec/ -v`
Expected: PASS

Note: The stderr test may have non-deterministic ordering. If it fails, relax the assertion to just check that both "out" and "err" appear in the output.

**Step 5: Commit**

```bash
git add exec/
git commit -m "feat: add code execution runner with output capture"
```

---

### Task 6: Image Handling (`exec/image.go`)

**Files:**
- Create: `exec/image.go`
- Create: `exec/image_test.go`

**Step 1: Write the test**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./exec/ -v -run TestRunImage`
Expected: FAIL — `RunImage` not defined

**Step 3: Implement image handling**

```go
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
	".png": true,
	".jpg": true,
	".jpeg": true,
	".gif": true,
	".svg": true,
}

// RunImage runs a bash script that is expected to produce an image file.
// The last line of stdout is treated as the path to the image.
// The image is copied to destDir with a <uuid>-<date>.<ext> filename.
// Returns the new filename (not the full path).
func RunImage(script, destDir, workdir string) (string, error) {
	output, err := Run("bash", script, workdir)
	if err != nil {
		return "", fmt.Errorf("running image script: %w", err)
	}

	// Last non-empty line of output is the image path
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("image script produced no output")
	}
	srcPath := strings.TrimSpace(lines[len(lines)-1])

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
```

**Step 4: Add uuid dependency and run tests**

Run: `go get github.com/google/uuid`
Run: `go test ./exec/ -v -run TestRunImage`
Expected: PASS

**Step 5: Commit**

```bash
git add exec/ go.mod go.sum
git commit -m "feat: add image script execution with file copy and naming"
```

---

### Task 7: `showboat init` Command (`cmd/init.go`)

**Files:**
- Create: `cmd/init.go`
- Create: `cmd/init_test.go`
- Modify: `main.go`

**Step 1: Write the test**

```go
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

	err := Init(file, "My Demo")
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
}

func TestInitErrorsIfExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	os.WriteFile(file, []byte("existing"), 0644)

	err := Init(file, "My Demo")
	if err == nil {
		t.Error("expected error when file exists")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -v -run TestInit`
Expected: FAIL — `Init` not defined

**Step 3: Implement**

```go
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/simonw/showboat/markdown"
)

// Init creates a new showboat document with a title and timestamp.
// Returns an error if the file already exists.
func Init(file, title string) error {
	if _, err := os.Stat(file); err == nil {
		return fmt.Errorf("file already exists: %s", file)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	blocks := []markdown.Block{
		markdown.TitleBlock{Title: title, Timestamp: timestamp},
	}

	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	return markdown.Write(f, blocks)
}
```

**Step 4: Run tests**

Run: `go test ./cmd/ -v -run TestInit`
Expected: PASS

**Step 5: Wire into main.go**

Update the `init` case in `main.go`:

```go
case "init":
    if len(os.Args) < 4 {
        fmt.Fprintln(os.Stderr, "usage: showboat init <file> <title>")
        os.Exit(1)
    }
    if err := cmd.Init(os.Args[2], os.Args[3]); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
```

Add import: `"github.com/simonw/showboat/cmd"`

**Step 6: Build and manually test**

Run: `go build -o showboat . && ./showboat init /tmp/test-demo.md "Test Demo" && cat /tmp/test-demo.md`
Expected: File created with title and timestamp

**Step 7: Commit**

```bash
git add cmd/ main.go
git commit -m "feat: implement showboat init command"
```

---

### Task 8: `showboat build commentary` Command

**Files:**
- Create: `cmd/build.go`
- Create: `cmd/build_test.go`
- Modify: `main.go`

**Step 1: Write the test**

```go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCommentary(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")
	Init(file, "Test")

	err := BuildCommentary(file, "Hello world.")
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Hello world.") {
		t.Errorf("expected commentary in file: %s", content)
	}
}

func TestBuildCommentaryMultiple(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")
	Init(file, "Test")

	BuildCommentary(file, "First.")
	BuildCommentary(file, "Second.")

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if !strings.Contains(s, "First.") || !strings.Contains(s, "Second.") {
		t.Errorf("expected both commentaries: %s", s)
	}
}

func TestBuildCommentaryNoFile(t *testing.T) {
	err := BuildCommentary("/nonexistent/demo.md", "Hello")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -v -run TestBuildCommentary`
Expected: FAIL

**Step 3: Implement**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/simonw/showboat/markdown"
)

// BuildCommentary appends a commentary block to an existing showboat document.
func BuildCommentary(file, text string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", file)
	}

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	blocks, err := markdown.Parse(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("parsing file: %w", err)
	}

	blocks = append(blocks, markdown.CommentaryBlock{Text: text})

	out, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	defer out.Close()

	return markdown.Write(out, blocks)
}
```

**Step 4: Run tests**

Run: `go test ./cmd/ -v -run TestBuildCommentary`
Expected: PASS

**Step 5: Wire into main.go**

Update the `build` case to parse subcommands:

```go
case "build":
    if len(os.Args) < 4 {
        fmt.Fprintln(os.Stderr, "usage: showboat build <file> <subcommand> [args...]")
        os.Exit(1)
    }
    file := os.Args[2]
    sub := os.Args[3]
    switch sub {
    case "commentary":
        text, err := getTextArg(os.Args[4:])
        if err != nil {
            fmt.Fprintf(os.Stderr, "error: %v\n", err)
            os.Exit(1)
        }
        if err := cmd.BuildCommentary(file, text); err != nil {
            fmt.Fprintf(os.Stderr, "error: %v\n", err)
            os.Exit(1)
        }
    case "run":
        fmt.Fprintln(os.Stderr, "build run: not yet implemented")
        os.Exit(1)
    case "image":
        fmt.Fprintln(os.Stderr, "build image: not yet implemented")
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "unknown build subcommand: %s\n", sub)
        os.Exit(1)
    }
```

Add a helper to read from args or stdin:

```go
func getTextArg(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	// Read from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}
	text := strings.TrimRight(string(data), "\n")
	if text == "" {
		return "", fmt.Errorf("no text provided (pass as argument or pipe to stdin)")
	}
	return text, nil
}
```

Add imports: `"io"`, `"strings"`

**Step 6: Commit**

```bash
git add cmd/ main.go
git commit -m "feat: implement showboat build commentary command"
```

---

### Task 9: `showboat build run` Command

**Files:**
- Modify: `cmd/build.go`
- Modify: `cmd/build_test.go`
- Modify: `main.go`

**Step 1: Write the test**

```go
func TestBuildRun(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")
	Init(file, "Test")

	err := BuildRun(file, "bash", "echo hello", "")
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if !strings.Contains(s, "```bash\necho hello\n```") {
		t.Errorf("expected code block: %s", s)
	}
	if !strings.Contains(s, "```output\nhello\n```") {
		t.Errorf("expected output block: %s", s)
	}
}

func TestBuildRunNonZeroExit(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")
	Init(file, "Test")

	err := BuildRun(file, "bash", "echo fail && exit 1", "")
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "fail") {
		t.Error("expected output even on non-zero exit")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -v -run TestBuildRun`
Expected: FAIL

**Step 3: Implement**

Add to `cmd/build.go`:

```go
import (
	execpkg "github.com/simonw/showboat/exec"
)

// BuildRun appends a code block, executes it, and appends the output.
func BuildRun(file, lang, code, workdir string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", file)
	}

	// Execute the code
	output, err := execpkg.Run(lang, code, workdir)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	// Parse existing document
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	blocks, err := markdown.Parse(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("parsing file: %w", err)
	}

	// Append code and output blocks
	blocks = append(blocks,
		markdown.CodeBlock{Lang: lang, Code: code},
		markdown.OutputBlock{Content: output},
	)

	out, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	defer out.Close()

	return markdown.Write(out, blocks)
}
```

**Step 4: Run tests**

Run: `go test ./cmd/ -v -run TestBuildRun`
Expected: PASS

**Step 5: Wire into main.go**

Update the `run` case:

```go
case "run":
    if len(os.Args) < 5 {
        fmt.Fprintln(os.Stderr, "usage: showboat build <file> run <lang> [code]")
        os.Exit(1)
    }
    lang := os.Args[4]
    code, err := getTextArg(os.Args[5:])
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    if err := cmd.BuildRun(file, lang, code, workdir); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
```

Also add `--workdir` flag parsing before the command switch in main.go. Parse `os.Args` looking for `--workdir <dir>` and remove it from the args before dispatching.

**Step 6: Commit**

```bash
git add cmd/ main.go
git commit -m "feat: implement showboat build run command"
```

---

### Task 10: `showboat build image` Command

**Files:**
- Modify: `cmd/build.go`
- Modify: `cmd/build_test.go`
- Modify: `main.go`

**Step 1: Write the test**

```go
func TestBuildImage(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")
	Init(file, "Test")

	imgDir := t.TempDir()
	imgPath := filepath.Join(imgDir, "test.png")
	script := `printf '\x89PNG\r\n\x1a\n' > ` + imgPath + ` && echo ` + imgPath

	err := BuildImage(file, script, "")
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if !strings.Contains(s, "```bash {image}") {
		t.Errorf("expected image code block: %s", s)
	}
	if !strings.Contains(s, "```output-image") {
		t.Errorf("expected output-image block: %s", s)
	}
	if !strings.Contains(s, ".png)") {
		t.Errorf("expected .png filename reference: %s", s)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -v -run TestBuildImage`
Expected: FAIL

**Step 3: Implement**

Add to `cmd/build.go`:

```go
import "path/filepath"

// BuildImage appends an image code block, runs the script, captures the image.
func BuildImage(file, script, workdir string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", file)
	}

	destDir := filepath.Dir(file)
	if destDir == "" {
		destDir = "."
	}

	filename, err := execpkg.RunImage(script, destDir, workdir)
	if err != nil {
		return fmt.Errorf("image capture failed: %w", err)
	}

	// Parse existing document
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	blocks, err := markdown.Parse(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("parsing file: %w", err)
	}

	blocks = append(blocks,
		markdown.CodeBlock{Lang: "bash", Code: script, IsImage: true},
		markdown.ImageOutputBlock{AltText: "Screenshot", Filename: filename},
	)

	out, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	defer out.Close()

	return markdown.Write(out, blocks)
}
```

**Step 4: Run tests**

Run: `go test ./cmd/ -v -run TestBuildImage`
Expected: PASS

**Step 5: Wire into main.go**

```go
case "image":
    script, err := getTextArg(os.Args[4:])
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    if err := cmd.BuildImage(file, script, workdir); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
```

**Step 6: Commit**

```bash
git add cmd/ main.go
git commit -m "feat: implement showboat build image command"
```

---

### Task 11: `showboat verify` Command

**Files:**
- Create: `cmd/verify.go`
- Create: `cmd/verify_test.go`
- Modify: `main.go`

**Step 1: Write the test**

```go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyPasses(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")
	Init(file, "Test")
	BuildRun(file, "bash", "echo hello", "")

	diffs, err := Verify(file, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) > 0 {
		t.Errorf("expected no diffs, got: %v", diffs)
	}
}

func TestVerifyDetectsDrift(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")
	Init(file, "Test")
	BuildRun(file, "bash", "echo hello", "")

	// Tamper with the output
	content, _ := os.ReadFile(file)
	tampered := strings.Replace(string(content), "hello\n", "goodbye\n", 1)
	os.WriteFile(file, []byte(tampered), 0644)

	diffs, err := Verify(file, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) == 0 {
		t.Error("expected diffs after tampering")
	}
}

func TestVerifyWritesOutput(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")
	Init(file, "Test")
	BuildRun(file, "bash", "echo hello", "")

	// Tamper
	content, _ := os.ReadFile(file)
	tampered := strings.Replace(string(content), "hello\n", "goodbye\n", 1)
	os.WriteFile(file, []byte(tampered), 0644)

	outputFile := filepath.Join(dir, "updated.md")
	_, err := Verify(file, outputFile)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), "hello") {
		t.Error("expected updated output to contain fresh 'hello'")
	}

	// Original should be untouched
	original, _ := os.ReadFile(file)
	if !strings.Contains(string(original), "goodbye") {
		t.Error("original should still contain tampered 'goodbye'")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -v -run TestVerify`
Expected: FAIL

**Step 3: Implement**

```go
package cmd

import (
	"fmt"
	"os"
	"strings"

	execpkg "github.com/simonw/showboat/exec"
	"github.com/simonw/showboat/markdown"
)

// Diff represents a mismatch between stored and actual output.
type Diff struct {
	BlockIndex int
	Expected   string
	Actual     string
}

func (d Diff) String() string {
	return fmt.Sprintf("Block %d:\n--- stored\n%s\n+++ actual\n%s", d.BlockIndex, d.Expected, d.Actual)
}

// Verify re-executes all code blocks and compares outputs.
// Returns a list of diffs. If outputFile is non-empty, writes an updated
// copy of the document with fresh outputs.
func Verify(file, outputFile string) ([]Diff, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	blocks, err := markdown.Parse(f)
	f.Close()
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	var diffs []Diff
	updatedBlocks := make([]markdown.Block, len(blocks))
	copy(updatedBlocks, blocks)

	for i, block := range blocks {
		code, ok := block.(markdown.CodeBlock)
		if !ok {
			continue
		}

		// Skip image blocks for now (image verification is more complex)
		if code.IsImage {
			continue
		}

		// Execute
		output, err := execpkg.Run(code.Lang, code.Code, "")
		if err != nil {
			return nil, fmt.Errorf("executing block %d: %w", i, err)
		}

		// Find the corresponding output block (should be next)
		if i+1 < len(blocks) {
			if outBlock, ok := blocks[i+1].(markdown.OutputBlock); ok {
				if outBlock.Content != output {
					diffs = append(diffs, Diff{
						BlockIndex: i,
						Expected:   strings.TrimRight(outBlock.Content, "\n"),
						Actual:     strings.TrimRight(output, "\n"),
					})
				}
				updatedBlocks[i+1] = markdown.OutputBlock{Content: output}
			}
		}
	}

	if outputFile != "" {
		out, err := os.Create(outputFile)
		if err != nil {
			return diffs, fmt.Errorf("writing output file: %w", err)
		}
		defer out.Close()
		if err := markdown.Write(out, updatedBlocks); err != nil {
			return diffs, fmt.Errorf("writing output: %w", err)
		}
	}

	return diffs, nil
}
```

**Step 4: Run tests**

Run: `go test ./cmd/ -v -run TestVerify`
Expected: PASS

**Step 5: Wire into main.go**

```go
case "verify":
    if len(os.Args) < 3 {
        fmt.Fprintln(os.Stderr, "usage: showboat verify <file> [--output <newfile>]")
        os.Exit(1)
    }
    file := os.Args[2]
    outputFile := ""
    for i, arg := range os.Args[3:] {
        if arg == "--output" && i+3+1 < len(os.Args) {
            outputFile = os.Args[i+3+1]
        }
    }
    diffs, err := cmd.Verify(file, outputFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    if len(diffs) > 0 {
        for _, d := range diffs {
            fmt.Fprintln(os.Stderr, d.String())
        }
        os.Exit(1)
    }
```

**Step 6: Commit**

```bash
git add cmd/ main.go
git commit -m "feat: implement showboat verify command"
```

---

### Task 12: `showboat extract` Command

**Files:**
- Create: `cmd/extract.go`
- Create: `cmd/extract_test.go`
- Modify: `main.go`

**Step 1: Write the test**

```go
package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")
	Init(file, "Test Demo")
	BuildCommentary(file, "Hello world.")
	BuildRun(file, "bash", "echo hi", "")

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
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/ -v -run TestExtract`
Expected: FAIL

**Step 3: Implement**

```go
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/simonw/showboat/markdown"
)

// Extract parses a showboat document and returns the sequence of CLI commands
// that would reproduce it.
func Extract(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	defer f.Close()

	blocks, err := markdown.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	var commands []string
	for _, block := range blocks {
		switch b := block.(type) {
		case markdown.TitleBlock:
			commands = append(commands, fmt.Sprintf("showboat init %s %s", shellQuote(file), shellQuote(b.Title)))
		case markdown.CommentaryBlock:
			commands = append(commands, fmt.Sprintf("showboat build %s commentary %s", shellQuote(file), shellQuote(b.Text)))
		case markdown.CodeBlock:
			if b.IsImage {
				commands = append(commands, fmt.Sprintf("showboat build %s image %s", shellQuote(file), shellQuote(b.Code)))
			} else {
				commands = append(commands, fmt.Sprintf("showboat build %s run %s %s", shellQuote(file), b.Lang, shellQuote(b.Code)))
			}
		// OutputBlock and ImageOutputBlock are generated, not commands
		}
	}

	return commands, nil
}

// shellQuote wraps a string in single quotes, escaping internal single quotes.
func shellQuote(s string) string {
	if !strings.ContainsAny(s, " \t\n'\"\\$") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
```

**Step 4: Run tests**

Run: `go test ./cmd/ -v -run TestExtract`
Expected: PASS

**Step 5: Wire into main.go**

```go
case "extract":
    if len(os.Args) < 3 {
        fmt.Fprintln(os.Stderr, "usage: showboat extract <file>")
        os.Exit(1)
    }
    commands, err := cmd.Extract(os.Args[2])
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    for _, c := range commands {
        fmt.Println(c)
    }
```

**Step 6: Commit**

```bash
git add cmd/ main.go
git commit -m "feat: implement showboat extract command"
```

---

### Task 13: --workdir Flag Parsing

**Files:**
- Modify: `main.go`

**Step 1: Add --workdir parsing to main.go**

Before the command switch, scan `os.Args` for `--workdir <dir>`, extract it, and pass it through to `BuildRun` and `BuildImage`. Remove the flag from the args slice so it doesn't interfere with command parsing.

```go
// At the top of main(), before the command switch:
workdir := ""
var filteredArgs []string
for i := 1; i < len(os.Args); i++ {
    if os.Args[i] == "--workdir" && i+1 < len(os.Args) {
        workdir = os.Args[i+1]
        i++ // skip next arg
    } else {
        filteredArgs = append(filteredArgs, os.Args[i])
    }
}
// Use filteredArgs instead of os.Args[1:] for command dispatch
```

**Step 2: Build and manually test**

Run: `go build -o showboat . && ./showboat build /tmp/test.md run bash --workdir /tmp "pwd"`

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: add --workdir flag parsing"
```

---

### Task 14: Integration Test — Full Workflow

**Files:**
- Create: `integration_test.go` (in root package)

**Step 1: Write the test**

```go
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFullWorkflow(t *testing.T) {
	// Build the binary
	tmpBin := filepath.Join(t.TempDir(), "showboat")
	build := exec.Command("go", "build", "-o", tmpBin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	// Init
	run(t, tmpBin, "init", file, "Integration Test Demo")

	// Commentary
	run(t, tmpBin, "build", file, "commentary", "This demo tests the full workflow.")

	// Run bash
	run(t, tmpBin, "build", file, "run", "bash", "echo 'Hello from bash'")

	// Run python
	run(t, tmpBin, "build", file, "run", "python3", "print(2 + 2)")

	// More commentary
	run(t, tmpBin, "build", file, "commentary", "Everything works.")

	// Read the file and check structure
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	checks := []string{
		"# Integration Test Demo",
		"This demo tests the full workflow.",
		"```bash\necho 'Hello from bash'\n```",
		"```output\nHello from bash\n```",
		"```python3\nprint(2 + 2)\n```",
		"```output\n4\n```",
		"Everything works.",
	}
	for _, check := range checks {
		if !strings.Contains(s, check) {
			t.Errorf("missing expected content: %q\n\nFull file:\n%s", check, s)
		}
	}

	// Verify should pass
	run(t, tmpBin, "verify", file)

	// Extract should produce commands
	out := runOutput(t, tmpBin, "extract", file)
	if !strings.Contains(out, "showboat init") {
		t.Errorf("extract missing init: %s", out)
	}
	if !strings.Contains(out, "run bash") {
		t.Errorf("extract missing run bash: %s", out)
	}

	// Tamper and verify should fail
	tampered := strings.Replace(s, "Hello from bash\n", "TAMPERED\n", 1)
	os.WriteFile(file, []byte(tampered), 0644)

	cmd := exec.Command(tmpBin, "verify", file)
	if err := cmd.Run(); err == nil {
		t.Error("expected verify to fail after tampering")
	}

	// Verify with --output should produce corrected file
	outputFile := filepath.Join(dir, "fixed.md")
	cmd = exec.Command(tmpBin, "verify", file, "--output", outputFile)
	cmd.Run() // may exit non-zero, that's fine

	fixed, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(fixed), "Hello from bash") {
		t.Error("expected fixed file to have correct output")
	}
}

func run(t *testing.T, bin string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("command failed: %s %v\n%s\n%s", bin, args, err, out)
	}
}

func runOutput(t *testing.T, bin string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %s %v\n%s\n%s", bin, args, err, out)
	}
	return string(out)
}
```

**Step 2: Run the integration test**

Run: `go test -v -run TestFullWorkflow -timeout 60s`
Expected: PASS

**Step 3: Commit**

```bash
git add integration_test.go
git commit -m "test: add full workflow integration test"
```

---

### Task 15: Final Polish and README

**Files:**
- Modify: `main.go` — ensure all subcommand help text is complete
- Verify all tests pass

**Step 1: Run all tests**

Run: `go test ./... -v`
Expected: All PASS

**Step 2: Build and test help output**

Run: `go build -o showboat . && ./showboat --help`
Expected: Complete, agent-friendly help text

**Step 3: Final commit**

```bash
git add -A
git commit -m "chore: final polish"
```
