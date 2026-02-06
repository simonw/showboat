# showcase

Create executable demo documents that show and prove an agent's work.

Showcase helps agents build markdown documents that mix commentary, executable code blocks, and captured output. These documents serve as both readable documentation and reproducible proof of work. A verifier can re-execute all code blocks and confirm the outputs still match.

## Usage

```
showcase init <file> <title>             Create a new demo document
showcase build <file> commentary [text]  Append commentary (text or stdin)
showcase build <file> run <lang> [code]  Run code and capture output
showcase build <file> image [script]     Run script, capture image output
showcase verify <file> [--output <new>]  Re-run and diff all code blocks
showcase extract <file>                  Emit build commands to recreate file
```

Build subcommands accept input from stdin when the text/code argument is omitted:

```bash
echo "Hello world" | showcase build demo.md commentary
cat script.sh | showcase build demo.md run bash
```

## Global options

- `--workdir <dir>` — Set working directory for code execution (default: current)
- `--help, -h` — Show help message

## Example

```bash
# Create a demo
showcase init demo.md "Setting Up a Python Project"

# Add commentary
showcase build demo.md commentary "First, let's create a virtual environment."

# Run a command and capture output
showcase build demo.md run bash "python3 -m venv .venv && echo 'Done'"

# Run Python and capture output
showcase build demo.md run python "print('Hello from Python')"

# Capture a screenshot
showcase build demo.md image "python screenshot.py http://localhost:8000"

# Verify the demo still works
showcase verify demo.md

# See what commands built the demo
showcase extract demo.md
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
