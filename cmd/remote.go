package cmd

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/simonw/showboat/markdown"
)

var remoteClient = &http.Client{Timeout: 10 * time.Second}

// documentID extracts the DocumentID from the first block if it is a TitleBlock.
func documentID(blocks []markdown.Block) string {
	if len(blocks) == 0 {
		return ""
	}
	tb, ok := blocks[0].(markdown.TitleBlock)
	if !ok {
		return ""
	}
	return tb.DocumentID
}

// postSection renders blocks to markdown and POSTs them form-encoded to
// SHOWBOAT_REMOTE_URL. No-op if the env var is unset or empty.
// Errors print a warning to stderr but do not fail the command.
func postSection(uuid, command string, blocks []markdown.Block) {
	remoteURL := os.Getenv("SHOWBOAT_REMOTE_URL")
	if remoteURL == "" {
		return
	}

	data := url.Values{}
	data.Set("uuid", uuid)
	data.Set("command", command)

	switch command {
	case "init":
		for _, b := range blocks {
			if tb, ok := b.(markdown.TitleBlock); ok {
				data.Set("title", tb.Title)
				break
			}
		}
	case "note":
		var buf strings.Builder
		markdown.Write(&buf, blocks)
		data.Set("markdown", buf.String())
	case "exec":
		for _, b := range blocks {
			switch blk := b.(type) {
			case markdown.CodeBlock:
				data.Set("language", blk.Lang)
				data.Set("input", blk.Code)
			case markdown.OutputBlock:
				data.Set("output", blk.Content)
			}
		}
	}

	resp, err := remoteClient.PostForm(remoteURL, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "showboat: remote POST warning: %v\n", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "showboat: remote POST warning: server returned %d\n", resp.StatusCode)
	}
}

// postImage POSTs an image as multipart/form-data to SHOWBOAT_REMOTE_URL.
// No-op if the env var is unset or empty.
func postImage(uuid string, blocks []markdown.Block, imagePath string) {
	remoteURL := os.Getenv("SHOWBOAT_REMOTE_URL")
	if remoteURL == "" {
		return
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	writer.WriteField("uuid", uuid)
	writer.WriteField("command", "image")

	for _, b := range blocks {
		if blk, ok := b.(markdown.ImageOutputBlock); ok {
			writer.WriteField("filename", blk.Filename)
			writer.WriteField("alt", blk.AltText)
		}
	}

	f, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "showboat: remote POST warning: %v\n", err)
		return
	}
	defer f.Close()

	part, err := writer.CreateFormFile("image", imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "showboat: remote POST warning: %v\n", err)
		return
	}
	io.Copy(part, f)
	writer.Close()

	resp, err := remoteClient.Post(remoteURL, writer.FormDataContentType(), &body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "showboat: remote POST warning: %v\n", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "showboat: remote POST warning: server returned %d\n", resp.StatusCode)
	}
}

// postPop POSTs a pop command to SHOWBOAT_REMOTE_URL.
// No-op if the env var is unset or empty.
func postPop(uuid string) {
	remoteURL := os.Getenv("SHOWBOAT_REMOTE_URL")
	if remoteURL == "" {
		return
	}

	data := url.Values{}
	data.Set("uuid", uuid)
	data.Set("command", "pop")

	resp, err := remoteClient.PostForm(remoteURL, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "showboat: remote POST warning: %v\n", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "showboat: remote POST warning: server returned %d\n", resp.StatusCode)
	}
}
