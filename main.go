package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"

	"github.com/simonw/showboat/cmd"
)

//go:embed help.txt
var helpText string

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
		if err := cmd.Init(args[1], args[2], version); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "note":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: showboat note <file> [text]")
			os.Exit(1)
		}
		text, err := getTextArg(args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err := cmd.Note(args[1], text); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "exec":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: showboat exec <file> <lang> [--filter <cmd>] [code]")
			os.Exit(1)
		}
		execFile := args[1]
		execLang := args[2]
		execFilter := ""
		var execRest []string
		for i := 3; i < len(args); i++ {
			if args[i] == "--filter" && i+1 < len(args) {
				execFilter = args[i+1]
				i++ // skip the filter value
			} else {
				execRest = append(execRest, args[i])
			}
		}
		code, err := getTextArg(execRest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		output, exitCode, err := cmd.Exec(execFile, execLang, code, execFilter, workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(output)
		if exitCode != 0 {
			os.Exit(exitCode)
		}

	case "image":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: showboat image <file> <image|![alt](image)>")
			os.Exit(1)
		}
		input, err := getTextArg(args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err := cmd.Image(args[1], input, workdir); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
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

	case "pop":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: showboat pop <file>")
			os.Exit(1)
		}
		if err := cmd.Pop(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
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
	fmt.Print(helpText)
}
