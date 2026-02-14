package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	execpkg "github.com/simonw/showboat/exec"
	"github.com/simonw/showboat/markdown"
)

// Diff represents a mismatch between expected and actual output of a code block.
type Diff struct {
	BlockIndex int
	Expected   string
	Actual     string
}

// String returns a human-readable description of the diff.
func (d Diff) String() string {
	return fmt.Sprintf("block %d:\n  expected: %s\n  actual:   %s",
		d.BlockIndex,
		strings.TrimRight(d.Expected, "\n"),
		strings.TrimRight(d.Actual, "\n"),
	)
}

// Verify re-executes all code blocks and compares outputs.
// If outputFile is non-empty, an updated copy of the document is written there.
// If workdir is non-empty, code blocks are executed in that directory.
func Verify(file, outputFile, workdir string) ([]Diff, error) {
	return VerifyWithPort(file, outputFile, workdir, 0)
}

// VerifyWithPort re-executes all code blocks and compares outputs.
// Server blocks are started as background processes with the given waitPort
// (or an auto-assigned port if waitPort is 0). The PORT environment variable
// is set for all blocks.
func VerifyWithPort(file, outputFile, workdir string, waitPort int) ([]Diff, error) {
	blocks, err := readBlocks(file)
	if err != nil {
		return nil, err
	}

	var diffs []Diff
	var servers []*execpkg.ServerProcess
	defer func() {
		for _, s := range servers {
			s.Stop()
		}
	}()

	// Port for $PORT env var — assigned when a server block is found
	port := waitPort

	for i := 0; i < len(blocks); i++ {
		cb, ok := blocks[i].(markdown.CodeBlock)
		if !ok || cb.IsImage {
			continue
		}

		if cb.IsServer {
			// Assign a port if not already set
			if port == 0 {
				p, err := execpkg.FreePort()
				if err != nil {
					return nil, fmt.Errorf("finding free port for server block %d: %w", i, err)
				}
				port = p
			}
			proc, err := execpkg.RunServer(cb.Lang, cb.Code, workdir, port, 10*time.Second)
			if err != nil {
				return nil, fmt.Errorf("starting server block %d: %w", i, err)
			}
			servers = append(servers, proc)
			continue
		}

		// Execute normal code block, with PORT in the environment if set
		output, _, err := runWithPort(cb.Lang, cb.Code, workdir, port)
		if err != nil {
			return nil, fmt.Errorf("executing block %d: %w", i, err)
		}

		// Check if next block is an OutputBlock
		if i+1 < len(blocks) {
			if ob, ok := blocks[i+1].(markdown.OutputBlock); ok {
				if ob.Content != output {
					diffs = append(diffs, Diff{
						BlockIndex: i,
						Expected:   ob.Content,
						Actual:     output,
					})
					// Update the block for the output copy
					blocks[i+1] = markdown.OutputBlock{Content: output}
				}
			}
		}
	}

	if outputFile != "" {
		if err := writeBlocks(outputFile, blocks); err != nil {
			return diffs, fmt.Errorf("writing output file: %w", err)
		}
	}

	return diffs, nil
}

// runWithPort executes code with the PORT environment variable set.
func runWithPort(lang, code, workdir string, port int) (string, int, error) {
	if port == 0 {
		return execpkg.Run(lang, code, workdir)
	}
	return execpkg.RunWithEnv(lang, code, workdir, []string{"PORT=" + strconv.Itoa(port)})
}

// Server starts a server process from a showboat document. It finds the first
// {server} code block, starts it, and prints the port. If waitPort is non-zero,
// it uses that port; otherwise it auto-assigns one.
func Server(file, workdir string, waitPort int) (*execpkg.ServerProcess, error) {
	blocks, err := readBlocks(file)
	if err != nil {
		return nil, err
	}

	for _, block := range blocks {
		cb, ok := block.(markdown.CodeBlock)
		if !ok || !cb.IsServer {
			continue
		}

		port := waitPort
		if port == 0 {
			p, err := execpkg.FreePort()
			if err != nil {
				return nil, fmt.Errorf("finding free port: %w", err)
			}
			port = p
		}

		// Set PORT so subsequent commands can use it
		os.Setenv("PORT", strconv.Itoa(port))

		proc, err := execpkg.RunServer(cb.Lang, cb.Code, workdir, port, 10*time.Second)
		if err != nil {
			return nil, fmt.Errorf("starting server: %w", err)
		}
		return proc, nil
	}

	return nil, fmt.Errorf("no {server} block found in %s", file)
}
