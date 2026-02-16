package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/simonw/showboat/markdown"
)

func TestPostSectionNoOpWhenEnvUnset(t *testing.T) {
	// Ensure SHOWBOAT_REMOTE_URL is not set
	t.Setenv("SHOWBOAT_REMOTE_URL", "")

	blocks := []markdown.Block{
		markdown.TitleBlock{Title: "Test", Timestamp: "2026-02-06T00:00:00Z", DocumentID: "test-uuid"},
	}

	// Should not panic or error
	postSection("test-uuid", "init", blocks)
}

func TestPostSectionSendsCorrectPayload(t *testing.T) {
	var gotBody string
	var gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("SHOWBOAT_REMOTE_URL", server.URL)

	blocks := []markdown.Block{
		markdown.TitleBlock{Title: "Test", Timestamp: "2026-02-06T00:00:00Z", DocumentID: "test-uuid"},
	}

	postSection("test-uuid", "init", blocks)

	if !strings.Contains(gotContentType, "application/x-www-form-urlencoded") {
		t.Errorf("expected form-urlencoded content type, got %q", gotContentType)
	}
	if !strings.Contains(gotBody, "uuid=test-uuid") {
		t.Errorf("expected uuid in body, got %q", gotBody)
	}
	if !strings.Contains(gotBody, "command=init") {
		t.Errorf("expected command=init in body, got %q", gotBody)
	}
}

func TestPostSectionNotePayload(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("SHOWBOAT_REMOTE_URL", server.URL)

	blocks := []markdown.Block{
		markdown.CommentaryBlock{Text: "Hello world."},
	}

	postSection("test-uuid", "note", blocks)

	if !strings.Contains(gotBody, "command=note") {
		t.Errorf("expected command=note in body, got %q", gotBody)
	}
	if !strings.Contains(gotBody, "markdown=") {
		t.Errorf("expected markdown field in body, got %q", gotBody)
	}
}

func TestPostSectionExecPayload(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("SHOWBOAT_REMOTE_URL", server.URL)

	blocks := []markdown.Block{
		markdown.CodeBlock{Lang: "bash", Code: "echo hello"},
		markdown.OutputBlock{Content: "hello\n"},
	}

	postSection("test-uuid", "exec", blocks)

	if !strings.Contains(gotBody, "command=exec") {
		t.Errorf("expected command=exec in body, got %q", gotBody)
	}
	if !strings.Contains(gotBody, "language=bash") {
		t.Errorf("expected language=bash in body, got %q", gotBody)
	}
}

func TestPostSectionServerErrorDoesNotPropagate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	t.Setenv("SHOWBOAT_REMOTE_URL", server.URL)

	blocks := []markdown.Block{
		markdown.TitleBlock{Title: "Test", Timestamp: "2026-02-06T00:00:00Z", DocumentID: "test-uuid"},
	}

	// Should not panic — errors are warnings only
	postSection("test-uuid", "init", blocks)
}

func TestPostSectionConnectionRefusedDoesNotPanic(t *testing.T) {
	// Point to a URL that will refuse connections
	t.Setenv("SHOWBOAT_REMOTE_URL", "http://127.0.0.1:1")

	blocks := []markdown.Block{
		markdown.TitleBlock{Title: "Test", Timestamp: "2026-02-06T00:00:00Z", DocumentID: "test-uuid"},
	}

	// Should not panic
	postSection("test-uuid", "init", blocks)
}

func TestPostPopSendsCorrectPayload(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("SHOWBOAT_REMOTE_URL", server.URL)

	postPop("test-uuid")

	if !strings.Contains(gotBody, "uuid=test-uuid") {
		t.Errorf("expected uuid in body, got %q", gotBody)
	}
	if !strings.Contains(gotBody, "command=pop") {
		t.Errorf("expected command=pop in body, got %q", gotBody)
	}
}

func TestPostPopNoOpWhenEnvUnset(t *testing.T) {
	t.Setenv("SHOWBOAT_REMOTE_URL", "")

	// Should not panic or error
	postPop("test-uuid")
}

func TestPostImageSendsMultipart(t *testing.T) {
	var gotContentType string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("SHOWBOAT_REMOTE_URL", server.URL)

	blocks := []markdown.Block{
		markdown.CodeBlock{Lang: "bash", Code: "python screenshot.py", IsImage: true},
		markdown.ImageOutputBlock{AltText: "Screenshot", Filename: "test.png"},
	}

	// Create a temporary image file for the test
	dir := t.TempDir()
	imgPath := dir + "/test.png"
	if err := writeTestFile(imgPath); err != nil {
		t.Fatal(err)
	}

	postImage("test-uuid", blocks, imgPath)

	if !strings.Contains(gotContentType, "multipart/form-data") {
		t.Errorf("expected multipart/form-data content type, got %q", gotContentType)
	}
	if !strings.Contains(gotBody, "test-uuid") {
		t.Errorf("expected uuid in body, got %q", gotBody)
	}
	if !strings.Contains(gotBody, "image") {
		t.Errorf("expected command=image in body, got %q", gotBody)
	}
	if !strings.Contains(gotBody, "filename") {
		t.Errorf("expected filename field in body, got %q", gotBody)
	}
	if strings.Contains(gotBody, "\"input\"") || strings.Contains(gotBody, "name=\"input\"") {
		t.Errorf("expected no input field in body, got %q", gotBody)
	}
}

func TestDocumentID(t *testing.T) {
	blocks := []markdown.Block{
		markdown.TitleBlock{Title: "Test", Timestamp: "2026-02-06T00:00:00Z", DocumentID: "my-uuid"},
		markdown.CommentaryBlock{Text: "Hello"},
	}

	id := documentID(blocks)
	if id != "my-uuid" {
		t.Errorf("expected 'my-uuid', got %q", id)
	}
}

func TestDocumentIDEmpty(t *testing.T) {
	blocks := []markdown.Block{
		markdown.CommentaryBlock{Text: "Hello"},
	}

	id := documentID(blocks)
	if id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}

func writeTestFile(path string) error {
	return writeTestFileWithContent(path, []byte("test content"))
}

func writeTestFileWithContent(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}
