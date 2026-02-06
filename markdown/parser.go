package markdown

import (
	"bufio"
	"io"
	"strings"
)

// Parse reads markdown from r and returns a slice of Blocks.
// The input is expected to be in the format produced by Write.
func Parse(r io.Reader) ([]Block, error) {
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var blocks []Block
	i := 0

	// skipSeparator consumes a single blank line between blocks.
	skipSeparator := func() {
		if i < len(lines) && lines[i] == "" {
			i++
		}
	}

	for i < len(lines) {
		// Title block: only at the very beginning of the document.
		if len(blocks) == 0 && strings.HasPrefix(lines[i], "# ") {
			title := lines[i][2:]
			i++ // past "# ..." line
			i++ // past blank line between title and timestamp
			// Parse timestamp: *timestamp*
			ts := strings.Trim(lines[i], "*")
			i++ // past "*timestamp*" line
			blocks = append(blocks, TitleBlock{Title: title, Timestamp: ts})
			skipSeparator()
			continue
		}

		// Fenced block: starts with ```
		if strings.HasPrefix(lines[i], "```") {
			fence := lines[i][3:]
			i++ // past opening fence

			switch {
			case fence == "output":
				var content strings.Builder
				for i < len(lines) && lines[i] != "```" {
					content.WriteString(lines[i])
					content.WriteString("\n")
					i++
				}
				i++ // past closing ```
				blocks = append(blocks, OutputBlock{Content: content.String()})

			case fence == "output-image":
				// Expect a single line: ![alt](filename)
				alt, filename := parseImageRef(lines[i])
				i++ // past image reference line
				i++ // past closing ```
				blocks = append(blocks, ImageOutputBlock{AltText: alt, Filename: filename})

			default:
				// Code block. Check for {image} suffix.
				lang := fence
				isImage := false
				if strings.HasSuffix(lang, " {image}") {
					lang = strings.TrimSuffix(lang, " {image}")
					isImage = true
				}
				var codeLines []string
				for i < len(lines) && lines[i] != "```" {
					codeLines = append(codeLines, lines[i])
					i++
				}
				i++ // past closing ```
				blocks = append(blocks, CodeBlock{
					Lang:    lang,
					Code:    strings.Join(codeLines, "\n"),
					IsImage: isImage,
				})
			}

			skipSeparator()
			continue
		}

		// Commentary block: accumulate lines until a fence or EOF.
		var textLines []string
		for i < len(lines) {
			if strings.HasPrefix(lines[i], "```") {
				break
			}
			textLines = append(textLines, lines[i])
			i++
		}
		// Trim trailing empty lines (they are inter-block separators, not content).
		for len(textLines) > 0 && textLines[len(textLines)-1] == "" {
			textLines = textLines[:len(textLines)-1]
		}
		if len(textLines) > 0 {
			blocks = append(blocks, CommentaryBlock{Text: strings.Join(textLines, "\n")})
		}
	}

	return blocks, nil
}

// parseImageRef extracts the alt text and filename from a markdown image
// reference of the form ![alt](filename).
func parseImageRef(line string) (alt, filename string) {
	start := strings.Index(line, "![")
	if start == -1 {
		return "", ""
	}
	rest := line[start+2:]
	closeBracket := strings.Index(rest, "](")
	if closeBracket == -1 {
		return "", ""
	}
	alt = rest[:closeBracket]
	rest = rest[closeBracket+2:]
	closeParen := strings.Index(rest, ")")
	if closeParen == -1 {
		return alt, ""
	}
	filename = rest[:closeParen]
	return alt, filename
}
