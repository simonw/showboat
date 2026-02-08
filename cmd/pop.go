package cmd

import (
	"fmt"

	"github.com/simonw/showboat/markdown"
)

// Pop removes the most recent entry from a showboat document.
// A "run" or "image" entry consists of a code block and its output, so both
// blocks are removed. A commentary entry is a single block.
// The title block cannot be removed.
func Pop(file string) error {
	blocks, err := readBlocks(file)
	if err != nil {
		return err
	}

	if len(blocks) == 0 {
		return fmt.Errorf("document is empty")
	}

	// Don't allow removing the title block.
	if len(blocks) == 1 {
		if _, ok := blocks[0].(markdown.TitleBlock); ok {
			return fmt.Errorf("nothing to pop: document only contains a title")
		}
	}

	last := blocks[len(blocks)-1]

	switch last.(type) {
	case markdown.OutputBlock, markdown.ImageOutputBlock:
		// Output blocks are always preceded by a code block — remove both.
		if len(blocks) >= 2 {
			blocks = blocks[:len(blocks)-2]
		} else {
			blocks = blocks[:len(blocks)-1]
		}
	default:
		blocks = blocks[:len(blocks)-1]
	}

	return writeBlocks(file, blocks)
}
