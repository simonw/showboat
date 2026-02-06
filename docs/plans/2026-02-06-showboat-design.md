# Showboat: Executable Demo Documents for Agents

## Overview

Showboat is a Go CLI tool for creating executable markdown demo documents. It helps agents produce documentation that both explains and proves their work — mixing commentary, executable code, and captured output in a single readable file.

The primary audience is developers evaluating agent work and non-technical stakeholders reviewing results. The markdown format is the source of truth: every CLI command maps 1:1 to a pattern in the markdown, making the format fully round-trippable.

Package: `github.com/simonw/showboat`

## Markdown Format

A showboat document is plain markdown with a strict structure. Each element maps to exactly one CLI command.

**Init (title + timestamp):**

```markdown
# My Demo Title

*2026-02-06T15:30:00Z*
```

**Commentary (prose):**

```markdown
Here's what we're doing next...
```

**Executed code with captured output:**

````markdown
```bash
curl https://api.example.com/health
```

```output
{"status": "ok"}
```
````

**Image capture:**

````markdown
```bash {image}
python screenshot.py http://localhost:8000
```

```output-image
![Description](a1b2c3d4-2026-02-06.png)
```
````

The `{image}` annotation on the code fence distinguishes an image-producing block from a regular execution block. The image file is copied to the same directory as the markdown file with a `<uuid>-<date>.<ext>` filename.

## CLI Commands

### `showboat init <file> <title>`

Creates a new demo file. Errors if the file already exists. Writes the title as an `# H1` heading and an ISO 8601 timestamp.

### `showboat build <file> commentary [text]`

Appends markdown prose. Text from argument or stdin.

### `showboat build <file> run <lang> [code]`

Appends a fenced code block tagged with `<lang>`, executes it, and appends the captured output in an `output` block. Code from argument or stdin.

### `showboat build <file> image [script]`

Appends a fenced `{image}` code block, executes the script, expects the last line of stdout to be a path to a newly created image file, copies the image to the markdown file's directory with a `<uuid>-<date>.<ext>` name, and appends an `output-image` block with the markdown image reference. Script from argument or stdin.

### `showboat verify <file> [--output <newfile>]`

Re-executes all code blocks in order, compares each output against what's stored in the document. Exits non-zero with a diff if anything changed. With `--output`, writes an updated copy to `<newfile>`.

### `showboat extract <file>`

Emits the sequence of `showboat init` and `showboat build` commands that would reproduce the document.

### Global Options

All execution commands accept `--workdir <dir>` to override the working directory (defaults to current directory).

All `build` subcommands read from stdin when the text/code argument is omitted.

## Help Text

The `showboat --help` output is the sole documentation for agents. It includes:

- A brief explanation of the tool's purpose: creating executable demo documents that show and prove an agent's work
- A complete list of subcommands with short descriptions
- A full worked example showing a realistic sequence: `init`, several `build` commands, then `verify` — including what the resulting markdown looks like
- The stdin convention

Each subcommand's `--help` shows its specific usage, arguments, and a short example. An agent should be able to run `showboat --help` once and immediately produce a correct demo.

## Project Structure

```
showboat/
├── main.go              # Entry point, top-level CLI routing
├── cmd/
│   ├── init.go          # showboat init
│   ├── build.go         # showboat build (commentary, run, image)
│   ├── verify.go        # showboat verify
│   └── extract.go       # showboat extract
├── markdown/
│   ├── parser.go        # Parse a showboat markdown file into structured blocks
│   ├── writer.go        # Serialize structured blocks back to markdown
│   └── blocks.go        # Block types: commentary, code, output, output-image
├── exec/
│   ├── runner.go        # Execute code blocks, capture stdout/stderr
│   └── image.go         # Run image scripts, copy output file, generate filename
├── go.mod
└── go.sum
```

The `markdown` package is the core — it owns the canonical representation. Both `build` (appending) and `extract` (reading back commands) operate through it. `verify` parses the file, re-runs via `exec`, and diffs the results.

## Execution & Error Handling

### Code Execution (`build run` and `verify`)

- Commands run via `os/exec` with the specified language as the interpreter (e.g., `bash -c "..."`, `python -c "..."`)
- Stdout and stderr are both captured and combined into the output block
- Non-zero exit codes don't stop the demo — agents may want to demonstrate error cases. The exit code is not recorded in the markdown; the output speaks for itself.

### Image Execution (`build image`)

- Same execution model, but the last line of stdout is treated as the image file path
- The tool verifies the file exists and is a recognized image format (png, jpg, gif, svg)
- Copies to `<uuid>-<date>.<ext>` in the same directory as the markdown file
- If the script fails or the file doesn't exist, the command errors and nothing is appended

### Verify Behavior

- Parses the document, re-executes every code block in order
- Compares each new output against the stored output (string comparison, or byte comparison for images)
- On mismatch: collects all diffs, prints them, exits non-zero
- With `--output`: writes the updated document to the specified file regardless of whether there are diffs
- Image verification: re-runs the image script, compares the new image to the stored one

### General Errors

- `init` on an existing file: error
- `build` on a non-existent file: error
- Missing stdin when no argument provided: error with clear message

## Testing Strategy

### Unit Tests (`markdown` package)

- Round-trip tests: parse a markdown document into blocks, serialize back, confirm identical output
- Test each block type in isolation: commentary, code+output, code+output-image
- Test `extract`: parse a document and verify the generated `showboat` commands would recreate it

### Unit Tests (`exec` package)

- Run simple bash/python commands, verify captured output
- Test image scripts with a trivial script that creates a small PNG, verify file copy and naming
- Test non-zero exit codes are handled gracefully

### Integration Tests (CLI level)

- Script a full workflow: `init` → several `build` commands → `verify` passes
- Modify an output block by hand → `verify` fails with expected diff
- `verify --output` produces a corrected copy
- `extract` on the result produces commands that recreate the document
- Stdin variants of each build command

### Test Fixtures

- Known-good markdown files in `testdata/` for parser tests
- Small helper scripts for image tests (e.g., a bash one-liner that creates a 1x1 PNG)

Tests are runnable with `go test ./...` and require no external dependencies beyond bash and python being available.
