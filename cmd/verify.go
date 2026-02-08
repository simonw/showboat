package cmd

import (
	"fmt"
	"strings"

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
	blocks, err := readBlocks(file)
	if err != nil {
		return nil, err
	}

	var diffs []Diff

	for i := 0; i < len(blocks); i++ {
		cb, ok := blocks[i].(markdown.CodeBlock)
		if !ok || cb.IsImage {
			continue
		}

		// Execute the code block
		output, _, err := execpkg.Run(cb.Lang, cb.Code, workdir)
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
