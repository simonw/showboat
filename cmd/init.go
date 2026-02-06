package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/simonw/showboat/markdown"
)

// Init creates a new showboat document with a title and timestamp.
// Returns an error if the file already exists.
func Init(file, title string) error {
	if _, err := os.Stat(file); err == nil {
		return fmt.Errorf("file already exists: %s", file)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	blocks := []markdown.Block{
		markdown.TitleBlock{Title: title, Timestamp: timestamp},
	}

	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	return markdown.Write(f, blocks)
}
