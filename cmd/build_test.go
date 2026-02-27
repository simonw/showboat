package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNote(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	if err := Note(file, "Hello world"); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "Hello world") {
		t.Errorf("expected commentary in file, got: %s", content)
	}
}

func TestNoteMultiple(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	if err := Note(file, "First comment"); err != nil {
		t.Fatal(err)
	}
	if err := Note(file, "Second comment"); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "First comment") {
		t.Errorf("expected first commentary in file, got: %s", s)
	}
	if !strings.Contains(s, "Second comment") {
		t.Errorf("expected second commentary in file, got: %s", s)
	}
}

func TestNoteNoFile(t *testing.T) {
	err := Note("/nonexistent/path/demo.md", "Hello")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestExec(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	if _, _, err := Exec(ExecOpts{File: file, Lang: "bash", Code: "echo hello"}); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "```bash\necho hello\n```") {
		t.Errorf("expected code block in file, got: %s", s)
	}
	if !strings.Contains(s, "```output\nhello\n```") {
		t.Errorf("expected output block in file, got: %s", s)
	}
}

func TestExecWithFilter(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	if _, _, err := Exec(ExecOpts{File: file, Lang: "python", Code: "1 + 1", Filter: "bash -c 'echo result: $(cat)'"}); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "```python\n1 + 1\n```") {
		t.Errorf("expected python code block in file, got: %s", s)
	}
	if !strings.Contains(s, "```output\nresult: 1 + 1\n```") {
		t.Errorf("expected output block with filter output in file, got: %s", s)
	}
}

func TestExecNonZeroExit(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	if _, _, err := Exec(ExecOpts{File: file, Lang: "bash", Code: "echo failing && exit 1"}); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "```bash\necho failing && exit 1\n```") {
		t.Errorf("expected code block in file, got: %s", s)
	}
	if !strings.Contains(s, "```output\nfailing\n```") {
		t.Errorf("expected output block with captured output, got: %s", s)
	}
}

// minimalPNG is a valid 1x1 white PNG used in tests.
var minimalPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG signature
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
	0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
	0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
	0x00, 0x00, 0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc,
	0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
	0x44, 0xae, 0x42, 0x60, 0x82,
}

func TestImage(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	pngPath := filepath.Join(dir, "test.png")
	if err := os.WriteFile(pngPath, minimalPNG, 0644); err != nil {
		t.Fatal(err)
	}

	if err := Image(file, pngPath, ""); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "```bash {image}") {
		t.Errorf("expected image code block in file, got: %s", s)
	}
	if !strings.Contains(s, "![") {
		t.Errorf("expected image output in file, got: %s", s)
	}
}

func TestImageMarkdownRef(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	pngPath := filepath.Join(dir, "test.png")
	if err := os.WriteFile(pngPath, minimalPNG, 0644); err != nil {
		t.Fatal(err)
	}

	input := "![My screenshot](" + pngPath + ")"

	if err := Image(file, input, ""); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "![My screenshot](") {
		t.Errorf("expected alt text 'My screenshot' in image output, got: %s", s)
	}
	if !strings.Contains(s, "```bash {image}") {
		t.Errorf("expected image code block in file, got: %s", s)
	}
}

func TestNoteCallsRemotePost(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("SHOWBOAT_REMOTE_URL", server.URL)

	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	// Reset gotBody so we only check the note POST
	gotBody = ""
	if err := Note(file, "Hello world"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(gotBody, "command=note") {
		t.Errorf("expected command=note in remote POST body, got %q", gotBody)
	}
	if !strings.Contains(gotBody, "uuid=") {
		t.Errorf("expected uuid in remote POST body, got %q", gotBody)
	}
}

func TestExecCallsRemotePost(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("SHOWBOAT_REMOTE_URL", server.URL)

	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	gotBody = ""
	if _, _, err := Exec(ExecOpts{File: file, Lang: "bash", Code: "echo hello"}); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(gotBody, "command=exec") {
		t.Errorf("expected command=exec in remote POST body, got %q", gotBody)
	}
	if !strings.Contains(gotBody, "language=bash") {
		t.Errorf("expected language=bash in remote POST body, got %q", gotBody)
	}
}

func TestImageCallsRemotePost(t *testing.T) {
	var gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("SHOWBOAT_REMOTE_URL", server.URL)

	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	pngPath := filepath.Join(dir, "test.png")
	if err := os.WriteFile(pngPath, minimalPNG, 0644); err != nil {
		t.Fatal(err)
	}

	gotContentType = ""
	if err := Image(file, pngPath, ""); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(gotContentType, "multipart/form-data") {
		t.Errorf("expected multipart/form-data content type for image POST, got %q", gotContentType)
	}
}

func TestParseImageInput(t *testing.T) {
	tests := []struct {
		input   string
		path    string
		altText string
	}{
		{"/path/to/img.png", "/path/to/img.png", ""},
		{"![alt text](/path/to/img.png)", "/path/to/img.png", "alt text"},
		{"![](file.jpg)", "file.jpg", ""},
		{"![Screenshot of homepage](shot.png)", "shot.png", "Screenshot of homepage"},
		{"  ![padded](file.png)  ", "file.png", "padded"},
		{"not-markdown.png", "not-markdown.png", ""},
		{`\![escaped](file.png)`, "file.png", "escaped"},
		{`\![alt text](/path/to/img.png)`, "/path/to/img.png", "alt text"},
	}
	for _, tt := range tests {
		path, alt := parseImageInput(tt.input)
		if path != tt.path {
			t.Errorf("parseImageInput(%q): path = %q, want %q", tt.input, path, tt.path)
		}
		if alt != tt.altText {
			t.Errorf("parseImageInput(%q): altText = %q, want %q", tt.input, alt, tt.altText)
		}
	}
}

func TestImageMarkdownRefEscapedBang(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	pngPath := filepath.Join(dir, "test.png")
	if err := os.WriteFile(pngPath, minimalPNG, 0644); err != nil {
		t.Fatal(err)
	}

	input := `\![My screenshot](` + pngPath + ")"

	if err := Image(file, input, ""); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "![My screenshot](") {
		t.Errorf("expected alt text 'My screenshot' in image output, got: %s", s)
	}
}

func TestImageMarkdownRefBadPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.md")

	if err := Init(file, "Test", "dev"); err != nil {
		t.Fatal(err)
	}

	input := "![alt text](/nonexistent/image.png)"
	err := Image(file, input, "")
	if err == nil {
		t.Error("expected error for nonexistent image path in markdown ref")
	}
}
