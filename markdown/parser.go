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
			// Skip blank line between title and timestamp
			if i < len(lines) && lines[i] == "" {
				i++
			}
			// Parse timestamp: *timestamp* or *timestamp by Showboat version*
			ts := ""
			ver := ""
			if i < len(lines) && strings.HasPrefix(lines[i], "*") && strings.HasSuffix(lines[i], "*") {
				dateline := strings.Trim(lines[i], "*")
				if idx := strings.Index(dateline, " by Showboat "); idx != -1 {
					ts = dateline[:idx]
					ver = dateline[idx+len(" by Showboat "):]
				} else {
					ts = dateline
				}
				i++
			}
			// Check for optional document ID comment after timestamp.
			docID := ""
			if i < len(lines) && strings.HasPrefix(lines[i], "<!-- showboat-id: ") && strings.HasSuffix(lines[i], " -->") {
				docID = strings.TrimPrefix(lines[i], "<!-- showboat-id: ")
				docID = strings.TrimSuffix(docID, " -->")
				i++
			}
			blocks = append(blocks, TitleBlock{Title: title, Timestamp: ts, Version: ver, DocumentID: docID})
			skipSeparator()
			continue
		}

		// Fenced block: starts with ``` (possibly more backticks)
		if strings.HasPrefix(lines[i], "```") {
			// Count the backticks in the opening fence.
			fenceTicks := 0
			for _, ch := range lines[i] {
				if ch == '`' {
					fenceTicks++
				} else {
					break
				}
			}
			closingFence := strings.Repeat("`", fenceTicks)
			info := lines[i][fenceTicks:]
			i++ // past opening fence

			switch {
			case info == "output":
				var content strings.Builder
				for i < len(lines) && lines[i] != closingFence {
					content.WriteString(lines[i])
					content.WriteString("\n")
					i++
				}
				i++ // past closing fence
				blocks = append(blocks, OutputBlock{Content: content.String()})

			default:
				// Code block. Extract {image} and {filter=...} suffixes.
				lang := info
				isImage := false
				filter := ""
				if strings.HasSuffix(lang, " {image}") {
					lang = strings.TrimSuffix(lang, " {image}")
					isImage = true
				}
				if idx := strings.Index(lang, " {filter="); idx != -1 {
					suffix := lang[idx+len(" {filter="):]
					if end := strings.Index(suffix, "}"); end != -1 {
						filter = suffix[:end]
						lang = lang[:idx] + suffix[end+1:]
					}
				}
				var codeLines []string
				for i < len(lines) && lines[i] != closingFence {
					codeLines = append(codeLines, lines[i])
					i++
				}
				i++ // past closing fence
				blocks = append(blocks, CodeBlock{
					Lang:    lang,
					Code:    strings.Join(codeLines, "\n"),
					IsImage: isImage,
					Filter:  filter,
				})
			}

			skipSeparator()
			continue
		}

		// Image output line: ![alt](filename) on its own line.
		if strings.HasPrefix(lines[i], "![") {
			alt, filename := parseImageRef(lines[i])
			if filename != "" {
				i++
				blocks = append(blocks, ImageOutputBlock{AltText: alt, Filename: filename})
				skipSeparator()
				continue
			}
		}

		// Commentary block: accumulate lines until a fence, image output, or EOF.
		var textLines []string
		for i < len(lines) {
			if strings.HasPrefix(lines[i], "```") {
				break
			}
			if strings.HasPrefix(lines[i], "![") {
				if _, fn := parseImageRef(lines[i]); fn != "" {
					break
				}
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
