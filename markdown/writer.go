package markdown

import (
	"fmt"
	"io"
)

// Write serializes a slice of Blocks to markdown, writing the result to w.
func Write(w io.Writer, blocks []Block) error {
	for i, block := range blocks {
		if i > 0 {
			if _, err := fmt.Fprint(w, "\n"); err != nil {
				return err
			}
		}
		if err := writeBlock(w, block); err != nil {
			return err
		}
	}
	return nil
}

func writeBlock(w io.Writer, block Block) error {
	switch b := block.(type) {
	case TitleBlock:
		_, err := fmt.Fprintf(w, "# %s\n\n*%s*\n", b.Title, b.Timestamp)
		return err
	case CommentaryBlock:
		_, err := fmt.Fprintf(w, "%s\n", b.Text)
		return err
	case CodeBlock:
		lang := b.Lang
		if b.IsImage {
			lang += " {image}"
		}
		_, err := fmt.Fprintf(w, "```%s\n%s\n```\n", lang, b.Code)
		return err
	case OutputBlock:
		_, err := fmt.Fprintf(w, "```output\n%s```\n", b.Content)
		return err
	case ImageOutputBlock:
		_, err := fmt.Fprintf(w, "![%s](%s)\n", b.AltText, b.Filename)
		return err
	default:
		return fmt.Errorf("unknown block type: %T", block)
	}
}
