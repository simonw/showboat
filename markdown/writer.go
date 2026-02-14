package markdown

import (
	"fmt"
	"io"
	"strings"
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
		dateline := b.Timestamp
		if b.Version != "" {
			dateline += " by Showboat " + b.Version
		}
		_, err := fmt.Fprintf(w, "# %s\n\n*%s*\n", b.Title, dateline)
		return err
	case CommentaryBlock:
		_, err := fmt.Fprintf(w, "%s\n", b.Text)
		return err
	case CodeBlock:
		lang := b.Lang
		if b.IsImage {
			lang += " {image}"
		} else if b.IsServer {
			lang += " {server}"
		}
		_, err := fmt.Fprintf(w, "```%s\n%s\n```\n", lang, b.Code)
		return err
	case OutputBlock:
		fence := fenceFor(b.Content)
		_, err := fmt.Fprintf(w, "%soutput\n%s%s\n", fence, b.Content, fence)
		return err
	case ImageOutputBlock:
		_, err := fmt.Fprintf(w, "![%s](%s)\n", b.AltText, b.Filename)
		return err
	default:
		return fmt.Errorf("unknown block type: %T", block)
	}
}

// fenceFor returns a backtick fence string (at least 3 backticks) that is
// longer than any backtick sequence found at the start of a line in content.
func fenceFor(content string) string {
	maxRun := 0
	for _, line := range strings.Split(content, "\n") {
		run := 0
		for _, ch := range line {
			if ch == '`' {
				run++
			} else {
				break
			}
		}
		if run > maxRun {
			maxRun = run
		}
	}
	n := 3
	if maxRun >= 3 {
		n = maxRun + 1
	}
	return strings.Repeat("`", n)
}
