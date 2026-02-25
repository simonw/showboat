package markdown

// Block is an element in a showboat document.
type Block interface {
	Type() string
}

// TitleBlock is the document header: an H1 title and a timestamp.
type TitleBlock struct {
	Title      string
	Timestamp  string
	Version    string
	DocumentID string
}

func (b TitleBlock) Type() string { return "title" }

// CommentaryBlock is free-form markdown prose.
type CommentaryBlock struct {
	Text string
}

func (b CommentaryBlock) Type() string { return "commentary" }

// CodeBlock is an executable fenced code block.
type CodeBlock struct {
	Lang    string
	Code    string
	IsImage bool
}

func (b CodeBlock) Type() string { return "code" }

// OutputBlock is captured text output from a code block.
// When Lang is non-empty the fence uses that language for syntax
// highlighting (e.g. ```go); otherwise it uses ```output.
type OutputBlock struct {
	Lang    string
	Content string
}

func (b OutputBlock) Type() string { return "output" }

// ImageOutputBlock is a captured image reference from an image code block.
type ImageOutputBlock struct {
	AltText  string
	Filename string
}

func (b ImageOutputBlock) Type() string { return "output-image" }
