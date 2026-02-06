package markdown

import "testing"

func TestCommentaryBlock(t *testing.T) {
	b := CommentaryBlock{Text: "Hello world\n\nThis is a test."}
	if b.Type() != "commentary" {
		t.Errorf("expected type commentary, got %s", b.Type())
	}
}

func TestCodeBlock(t *testing.T) {
	b := CodeBlock{Lang: "bash", Code: "echo hello", IsImage: false}
	if b.Type() != "code" {
		t.Errorf("expected type code, got %s", b.Type())
	}
}

func TestOutputBlock(t *testing.T) {
	b := OutputBlock{Content: "hello\n"}
	if b.Type() != "output" {
		t.Errorf("expected type output, got %s", b.Type())
	}
}

func TestImageOutputBlock(t *testing.T) {
	b := ImageOutputBlock{AltText: "Screenshot", Filename: "abc-2026-02-06.png"}
	if b.Type() != "output-image" {
		t.Errorf("expected type output-image, got %s", b.Type())
	}
}

func TestTitleBlock(t *testing.T) {
	b := TitleBlock{Title: "My Demo", Timestamp: "2026-02-06T15:30:00Z"}
	if b.Type() != "title" {
		t.Errorf("expected type title, got %s", b.Type())
	}
}
