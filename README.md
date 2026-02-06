# Showboat

Create executable demo documents that show and prove an agent's work.

Showboat helps agents build markdown documents that mix commentary, executable code blocks, and captured output. These documents serve as both readable documentation and reproducible proof of work. A verifier can re-execute all code blocks and confirm the outputs still match.

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
showboat build <file> commentary [text]  Append commentary (text or stdin)
showboat build <file> run <lang> [code]  Run code and capture output
showboat build <file> image [script]     Run script, capture image output
showboat verify <file> [--output <new>]  Re-run and diff all code blocks
showboat extract <file> [--filename <name>]  Emit build commands to recreate file
```

Build subcommands accept input from stdin when the text/code argument is omitted:

```bash
echo "Hello world" | showboat build demo.md commentary
cat script.sh | showboat build demo.md run bash
```

## Global options

- `--workdir <dir>` — Set working directory for code execution (default: current)
- `--version` — Print version and exit
- `--help, -h` — Show help message

## Example

```bash
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

`showboat extract` emits the sequence of `showboat init` and `showboat build` commands that would recreate a document from scratch:

```bash
showboat extract demo.md
```

For the example above this would output:

```
showboat init demo.md 'Setting Up a Python Project'
showboat build demo.md commentary 'First, let'\''s create a virtual environment.'
showboat build demo.md run bash 'python3 -m venv .venv && echo '\''Done'\'''
showboat build demo.md run python 'print('\''Hello from Python'\'')'
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