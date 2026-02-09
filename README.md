# Showboat

[![PyPI](https://img.shields.io/pypi/v/showboat.svg)](https://pypi.org/project/showboat/)
[![Changelog](https://img.shields.io/github/v/release/simonw/showboat?include_prereleases&label=changelog)](https://github.com/simonw/showboat/releases)
[![Tests](https://github.com/simonw/showboat/actions/workflows/test.yml/badge.svg)](https://github.com/simonw/showboat/actions/workflows/test.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://github.com/simonw/showboat/blob/main/LICENSE)

Create executable demo documents that show and prove an agent's work.

Showboat helps agents build markdown documents that mix commentary, executable code blocks, and captured output. These documents serve as both readable documentation and reproducible proof of work. A verifier can re-execute all code blocks and confirm the outputs still match.

## Example

Here's [an example Showboat demo document](https://github.com/simonw/showboat-demos/blob/main/shot-scraper/README.md) that demonstrates [shot-scraper](https://github.com/shot-scraper). It was created by Claude Code, as shown by [this transcript](https://gisthost.github.io/?29b0d0ebef50c57e7985a6004aad01c4/page-001.html#msg-2026-02-06T07-33-41-296Z).

## Installation

This Go tool can be installed directly [from PyPI](https://pypi.org/project/showboat/) using `pip` or `uv`.

You can run it without installing it first using `uvx`:

```bash
uvx showboat --help
```
Or install it like this, then run `showboat --help`:
```bash
uv tool install showboat
# or
pip install showboat
```

You can also install the Go binary directly:
```bash
go install github.com/simonw/showboat@latest
```
Or run it without installation like this:
```bash
go run github.com/simonw/showboat@latest --help
```
Compiled binaries are available [on the releases page](https://github.com/simonw/showboat/releases). On macOS you may need to [follow these extra steps](https://support.apple.com/en-us/102445) to use those.

## Usage

```
showboat init <file> <title>             Create a new demo document
showboat note <file> [text]              Append commentary (text or stdin)
showboat exec <file> <lang> [code]       Run code and capture output
showboat image <file> [script]           Run script, capture image output
showboat pop <file>                      Remove the most recent entry
showboat verify <file> [--output <new>]  Re-run and diff all code blocks
showboat extract <file> [--filename <name>]  Emit commands to recreate file
```

Commands accept input from stdin when the text/code argument is omitted:

```bash
echo "Hello world" | showboat note demo.md
cat script.sh | showboat exec demo.md bash
```

## Global options

- `--workdir <dir>` — Set working directory for code execution (default: current)
- `--version` — Print version and exit
- `--help, -h` — Show help message

## Exec output

The `exec` command prints the captured shell output to stdout and exits with the same exit code as the executed command. This lets agents see what happened during execution and react to errors. The output is still appended to the document regardless of exit code.

```bash
$ showboat exec demo.md bash "echo hello && exit 1"
hello
$ echo $?
1
```

## Popping entries

`showboat pop` removes the most recent entry from a document. For an `exec` or `image` entry this removes both the code block and its output. For a `note` entry it removes the single commentary block.

This is useful when a command produces an error that shouldn't remain in the document — the agent can inspect the output, decide the entry was a mistake, and pop it:

```bash
# A command fails
showboat exec demo.md bash "some-broken-command"

# Remove the failed entry from the document
showboat pop demo.md
```

## Example

```bash
# Create a demo
showboat init demo.md "Setting Up a Python Project"

# Add commentary
showboat note demo.md "First, let's create a virtual environment."

# Run a command and capture output
showboat exec demo.md bash "python3 -m venv .venv && echo 'Done'"

# Run Python and capture output
showboat exec demo.md python "print('Hello from Python')"

# Capture a screenshot
showboat image demo.md "python screenshot.py http://localhost:8000"
```

This produces a markdown file like:

````markdown
# Setting Up a Python Project

*2026-02-06T15:30:00Z*

First, let's create a virtual environment.

```bash
python3 -m venv .venv && echo 'Done'
```

```output
Done
```

```python
print('Hello from Python')
```

```output
Hello from Python
```
````

## Verifying

`showboat verify` re-executes every code block in a document and checks that the outputs still match:

```bash
showboat verify demo.md
```

## Extracting

`showboat extract` emits the sequence of commands that would recreate a document from scratch:

```bash
showboat extract demo.md
```

For the example above this would output:

```
showboat init demo.md 'Setting Up a Python Project'
showboat note demo.md 'First, let'\''s create a virtual environment.'
showboat exec demo.md bash 'python3 -m venv .venv && echo '\''Done'\'''
showboat exec demo.md python 'print('\''Hello from Python'\'')'
```

By default the commands reference the original filename. Use `--filename` to substitute a different filename in the emitted commands:

```bash
showboat extract demo.md --filename copy.md
```

## Building the Python wheels

The Python wheel versions are built using [go-to-wheel](https://github.com/simonw/go-to-wheel):

```bash
uvx go-to-wheel . \
  --readme README.md \
  --description "Create executable documents that demonstrate an agent's work" \
  --author 'Simon Willison' \
  --license Apache-2.0 \
  --url https://github.com/simonw/showboat \
  --set-version-var main.version \
  --version 0.1.0
```
