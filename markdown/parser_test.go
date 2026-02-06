package markdown

import (
	"strings"
	"testing"
)

func TestParseTitle(t *testing.T) {
	input := "# My Demo\n\n*2026-02-06T15:30:00Z*\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	tb, ok := blocks[0].(TitleBlock)
	if !ok {
		t.Fatalf("expected TitleBlock, got %T", blocks[0])
	}
	if tb.Title != "My Demo" {
		t.Errorf("expected title 'My Demo', got %q", tb.Title)
	}
	if tb.Timestamp != "2026-02-06T15:30:00Z" {
		t.Errorf("expected timestamp '2026-02-06T15:30:00Z', got %q", tb.Timestamp)
	}
}

func TestParseCommentary(t *testing.T) {
	input := "# Demo\n\n*2026-02-06T00:00:00Z*\n\nHello world.\n\nMore text here.\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d: %+v", len(blocks), blocks)
	}
	cb, ok := blocks[1].(CommentaryBlock)
	if !ok {
		t.Fatalf("expected CommentaryBlock, got %T", blocks[1])
	}
	if cb.Text != "Hello world.\n\nMore text here." {
		t.Errorf("unexpected text: %q", cb.Text)
	}
}

func TestParseCodeAndOutput(t *testing.T) {
	input := "```bash\necho hello\n```\n\n```output\nhello\n```\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	code, ok := blocks[0].(CodeBlock)
	if !ok {
		t.Fatalf("expected CodeBlock, got %T", blocks[0])
	}
	if code.Lang != "bash" || code.Code != "echo hello" || code.IsImage {
		t.Errorf("unexpected code block: %+v", code)
	}
	out, ok := blocks[1].(OutputBlock)
	if !ok {
		t.Fatalf("expected OutputBlock, got %T", blocks[1])
	}
	if out.Content != "hello\n" {
		t.Errorf("unexpected output: %q", out.Content)
	}
}

func TestParseImageCodeAndOutput(t *testing.T) {
	input := "```bash {image}\npython screenshot.py\n```\n\n```output-image\n![Screenshot](abc-2026-02-06.png)\n```\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	code, ok := blocks[0].(CodeBlock)
	if !ok {
		t.Fatalf("expected CodeBlock, got %T", blocks[0])
	}
	if !code.IsImage {
		t.Error("expected IsImage=true")
	}
	img, ok := blocks[1].(ImageOutputBlock)
	if !ok {
		t.Fatalf("expected ImageOutputBlock, got %T", blocks[1])
	}
	if img.AltText != "Screenshot" || img.Filename != "abc-2026-02-06.png" {
		t.Errorf("unexpected image output: %+v", img)
	}
}

func TestRoundTrip(t *testing.T) {
	input := "# Demo\n\n*2026-02-06T00:00:00Z*\n\nLet's begin.\n\n```bash\necho hi\n```\n\n```output\nhi\n```\n\nDone.\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	var buf strings.Builder
	err = Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != input {
		t.Errorf("round trip mismatch.\nexpected:\n%s\ngot:\n%s", input, buf.String())
	}
}
