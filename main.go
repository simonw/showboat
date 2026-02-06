package main

import (
	"fmt"
	"io"
	"os"

	"github.com/simonw/showboat/cmd"
)

var version = "dev"

func main() {
	args, workdir, showVersion := parseGlobalFlags(os.Args[1:])

	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "init":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: showboat init <file> <title>")
			os.Exit(1)
		}
		if err := cmd.Init(args[1], args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "build":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: showboat build <file> <subcommand> [args...]")
			os.Exit(1)
		}
		file := args[1]
		sub := args[2]
		remaining := args[3:]

		switch sub {
		case "commentary":
			text, err := getTextArg(remaining)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			if err := cmd.BuildCommentary(file, text); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "run":
			if len(remaining) < 1 {
				fmt.Fprintln(os.Stderr, "usage: showboat build <file> run <lang> [code]")
				os.Exit(1)
			}
			lang := remaining[0]
			code, err := getTextArg(remaining[1:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			if err := cmd.BuildRun(file, lang, code, workdir); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "image":
			script, err := getTextArg(remaining)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			if err := cmd.BuildImage(file, script, workdir); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "unknown build subcommand: %s\n", sub)
			os.Exit(1)
		}

	case "verify":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: showboat verify <file> [--output <new>]")
			os.Exit(1)
		}
		file := args[1]
		outputFile := ""
		remaining := args[2:]
		for i := 0; i < len(remaining); i++ {
			if remaining[i] == "--output" && i+1 < len(remaining) {
				outputFile = remaining[i+1]
				i++
			}
		}
		diffs, err := cmd.Verify(file, outputFile, workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(diffs) > 0 {
			for _, d := range diffs {
				fmt.Println(d.String())
			}
			os.Exit(1)
		}

	case "extract":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: showboat extract <file> [--filename <name>]")
			os.Exit(1)
		}
		extractFile := args[1]
		extractOutput := ""
		extractRemaining := args[2:]
		for i := 0; i < len(extractRemaining); i++ {
			if extractRemaining[i] == "--filename" && i+1 < len(extractRemaining) {
				extractOutput = extractRemaining[i+1]
				i++
			}
		}
		commands, err := cmd.Extract(extractFile, extractOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		for _, c := range commands {
			fmt.Println(c)
		}

	case "--help", "-h", "help":
		printUsage()
		os.Exit(0)

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

// parseGlobalFlags extracts global flags from args and returns the remaining
// args, workdir value, and whether to show version.
func parseGlobalFlags(args []string) (remaining []string, workdir string, showVersion bool) {
	for i := 0; i < len(args); i++ {
		if args[i] == "--workdir" && i+1 < len(args) {
			workdir = args[i+1]
			i++ // skip value
		} else if args[i] == "--version" {
			showVersion = true
		} else {
			remaining = append(remaining, args[i])
		}
	}
	return remaining, workdir, showVersion
}

// getTextArg returns args[0] if present, otherwise reads all of stdin.
func getTextArg(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}
	return string(data), nil
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
  showboat extract <file> [--filename <name>]  Emit build commands to recreate file

Global Options:
  --workdir <dir>   Set working directory for code execution (default: current)
  --version         Print version and exit
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
