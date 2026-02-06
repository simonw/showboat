package markdown

import (
	"strings"
	"testing"
)

func TestWriteTitle(t *testing.T) {
	var buf strings.Builder
	blocks := []Block{
		TitleBlock{Title: "My Demo", Timestamp: "2026-02-06T15:30:00Z"},
	}
	err := Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	expected := "# My Demo\n\n*2026-02-06T15:30:00Z*\n"
	if buf.String() != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, buf.String())
	}
}

func TestWriteCommentary(t *testing.T) {
	var buf strings.Builder
	blocks := []Block{
		CommentaryBlock{Text: "Hello world."},
	}
	err := Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Hello world.\n"
	if buf.String() != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, buf.String())
	}
}

func TestWriteCodeAndOutput(t *testing.T) {
	var buf strings.Builder
	blocks := []Block{
		CodeBlock{Lang: "bash", Code: "echo hello"},
		OutputBlock{Content: "hello\n"},
	}
	err := Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	expected := "```bash\necho hello\n```\n\n```output\nhello\n```\n"
	if buf.String() != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, buf.String())
	}
}

func TestWriteImageCodeAndOutput(t *testing.T) {
	var buf strings.Builder
	blocks := []Block{
		CodeBlock{Lang: "bash", Code: "python screenshot.py", IsImage: true},
		ImageOutputBlock{AltText: "Screenshot", Filename: "abc-2026-02-06.png"},
	}
	err := Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	expected := "```bash {image}\npython screenshot.py\n```\n\n![Screenshot](abc-2026-02-06.png)\n"
	if buf.String() != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, buf.String())
	}
}

func TestWriteFullDocument(t *testing.T) {
	var buf strings.Builder
	blocks := []Block{
		TitleBlock{Title: "Demo", Timestamp: "2026-02-06T00:00:00Z"},
		CommentaryBlock{Text: "Let's begin."},
		CodeBlock{Lang: "bash", Code: "echo hi"},
		OutputBlock{Content: "hi\n"},
		CommentaryBlock{Text: "Done."},
	}
	err := Write(&buf, blocks)
	if err != nil {
		t.Fatal(err)
	}
	expected := "# Demo\n\n*2026-02-06T00:00:00Z*\n\nLet's begin.\n\n```bash\necho hi\n```\n\n```output\nhi\n```\n\nDone.\n"
	if buf.String() != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, buf.String())
	}
}
