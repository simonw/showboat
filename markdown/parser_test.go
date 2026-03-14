package markdown

import (
	"strings"
	"testing"
)

func TestParseTitle(t *testing.T) {
	input := "# My Demo\n\n*2026-02-06T15:30:00Z by Showboat v0.3.0*\n"
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
	if tb.Version != "v0.3.0" {
		t.Errorf("expected version 'v0.3.0', got %q", tb.Version)
	}
}

func TestParseTitleNoVersion(t *testing.T) {
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
	if tb.Timestamp != "2026-02-06T15:30:00Z" {
		t.Errorf("expected timestamp '2026-02-06T15:30:00Z', got %q", tb.Timestamp)
	}
	if tb.Version != "" {
		t.Errorf("expected empty version, got %q", tb.Version)
	}
}

func TestParseCommentary(t *testing.T) {
	input := "# Demo\n\n*2026-02-06T00:00:00Z by Showboat v0.3.0*\n\nHello world.\n\nMore text here.\n"
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
	input := "```bash {image}\npython screenshot.py\n```\n\n![Screenshot](abc-2026-02-06.png)\n"
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

func TestParseOutputWithLongerFence(t *testing.T) {
	input := "````output\n```bash\necho hello\n```\n````\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d: %+v", len(blocks), blocks)
	}
	out, ok := blocks[0].(OutputBlock)
	if !ok {
		t.Fatalf("expected OutputBlock, got %T", blocks[0])
	}
	expected := "```bash\necho hello\n```\n"
	if out.Content != expected {
		t.Errorf("expected content:\n%s\ngot:\n%s", expected, out.Content)
	}
}

func TestParseCodeBlockWithLongerFence(t *testing.T) {
	input := "````bash\necho ```hello```\n````\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d: %+v", len(blocks), blocks)
	}
	code, ok := blocks[0].(CodeBlock)
	if !ok {
		t.Fatalf("expected CodeBlock, got %T", blocks[0])
	}
	if code.Code != "echo ```hello```" {
		t.Errorf("unexpected code: %q", code.Code)
	}
}

func TestRoundTripWithBackticksInOutput(t *testing.T) {
	input := "```bash\ncat inner.md\n```\n\n````output\n# My Demo\n\n```bash\necho hello\n```\n\n```output\nhello\n```\n````\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d: %+v", len(blocks), blocks)
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

func TestParseTitleWithDocumentID(t *testing.T) {
	input := "# My Demo\n\n*2026-02-06T15:30:00Z by Showboat v0.3.0*\n<!-- showboat-id: abc-123 -->\n"
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
	if tb.DocumentID != "abc-123" {
		t.Errorf("expected DocumentID 'abc-123', got %q", tb.DocumentID)
	}
	if tb.Title != "My Demo" {
		t.Errorf("expected title 'My Demo', got %q", tb.Title)
	}
	if tb.Timestamp != "2026-02-06T15:30:00Z" {
		t.Errorf("expected timestamp '2026-02-06T15:30:00Z', got %q", tb.Timestamp)
	}
}

func TestParseTitleWithoutDocumentID(t *testing.T) {
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
	if tb.DocumentID != "" {
		t.Errorf("expected empty DocumentID, got %q", tb.DocumentID)
	}
}

func TestParseTitleWithDocumentIDFollowedByContent(t *testing.T) {
	input := "# My Demo\n\n*2026-02-06T15:30:00Z*\n<!-- showboat-id: abc-123 -->\n\nHello world.\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d: %+v", len(blocks), blocks)
	}
	tb, ok := blocks[0].(TitleBlock)
	if !ok {
		t.Fatalf("expected TitleBlock, got %T", blocks[0])
	}
	if tb.DocumentID != "abc-123" {
		t.Errorf("expected DocumentID 'abc-123', got %q", tb.DocumentID)
	}
	cb, ok := blocks[1].(CommentaryBlock)
	if !ok {
		t.Fatalf("expected CommentaryBlock, got %T", blocks[1])
	}
	if cb.Text != "Hello world." {
		t.Errorf("unexpected text: %q", cb.Text)
	}
}

func TestRoundTripWithDocumentID(t *testing.T) {
	input := "# Demo\n\n*2026-02-06T00:00:00Z by Showboat v0.3.0*\n<!-- showboat-id: test-uuid-456 -->\n\nLet's begin.\n\n```bash\necho hi\n```\n\n```output\nhi\n```\n\nDone.\n"
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

func TestParseEmptyInput(t *testing.T) {
	blocks, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks for empty input, got %d", len(blocks))
	}
}

func TestParseCommentaryOnly(t *testing.T) {
	input := "Just some plain text.\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	cb, ok := blocks[0].(CommentaryBlock)
	if !ok {
		t.Fatalf("expected CommentaryBlock, got %T", blocks[0])
	}
	if cb.Text != "Just some plain text." {
		t.Errorf("unexpected text: %q", cb.Text)
	}
}

func TestParseMultipleCommentaryBlocksSeparatedByCode(t *testing.T) {
	input := "First paragraph.\n\n```bash\necho hi\n```\n\nSecond paragraph.\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d: %+v", len(blocks), blocks)
	}
	if _, ok := blocks[0].(CommentaryBlock); !ok {
		t.Errorf("expected CommentaryBlock at 0, got %T", blocks[0])
	}
	if _, ok := blocks[1].(CodeBlock); !ok {
		t.Errorf("expected CodeBlock at 1, got %T", blocks[1])
	}
	if _, ok := blocks[2].(CommentaryBlock); !ok {
		t.Errorf("expected CommentaryBlock at 2, got %T", blocks[2])
	}
}

func TestParseImageRefMalformed(t *testing.T) {
	// Missing closing paren
	alt, filename := parseImageRef("![alt](no-close-paren")
	if filename != "" {
		t.Errorf("expected empty filename for malformed ref, got %q", filename)
	}
	_ = alt

	// Missing ]( separator
	alt2, filename2 := parseImageRef("![alt text no bracket")
	if filename2 != "" {
		t.Errorf("expected empty filename, got %q", filename2)
	}
	_ = alt2

	// Not an image ref at all
	alt3, filename3 := parseImageRef("just plain text")
	if filename3 != "" {
		t.Errorf("expected empty filename for plain text, got %q", filename3)
	}
	_ = alt3
}

func TestParseCodeBlockEmptyContent(t *testing.T) {
	input := "```bash\n```\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	code, ok := blocks[0].(CodeBlock)
	if !ok {
		t.Fatalf("expected CodeBlock, got %T", blocks[0])
	}
	if code.Code != "" {
		t.Errorf("expected empty code, got %q", code.Code)
	}
}

func TestParseOutputBlockEmpty(t *testing.T) {
	input := "```output\n```\n"
	blocks, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	out, ok := blocks[0].(OutputBlock)
	if !ok {
		t.Fatalf("expected OutputBlock, got %T", blocks[0])
	}
	if out.Content != "" {
		t.Errorf("expected empty output, got %q", out.Content)
	}
}

func TestRoundTrip(t *testing.T) {
	input := "# Demo\n\n*2026-02-06T00:00:00Z by Showboat v0.3.0*\n\nLet's begin.\n\n```bash\necho hi\n```\n\n```output\nhi\n```\n\nDone.\n"
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
