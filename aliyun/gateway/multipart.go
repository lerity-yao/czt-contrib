package gateway

import (
	"bytes"
	"mime/multipart"
)

// MultipartBuilder helps construct multipart/form-data request bodies for DoRaw.
// It eliminates the boilerplate of buffer management and writer.Close() ordering.
type MultipartBuilder struct {
	writer *multipart.Writer
	buf    *bytes.Buffer
}

// NewMultipart creates a new MultipartBuilder.
func NewMultipart() *MultipartBuilder {
	buf := &bytes.Buffer{}
	return &MultipartBuilder{
		writer: multipart.NewWriter(buf),
		buf:    buf,
	}
}

// Field adds a text field to the multipart form.
func (b *MultipartBuilder) Field(name, value string) *MultipartBuilder {
	b.writer.WriteField(name, value)
	return b
}

// File adds a file field to the multipart form.
// Internally uses multipart.Writer.CreateFormFile.
func (b *MultipartBuilder) File(name, filename string, content []byte) *MultipartBuilder {
	part, _ := b.writer.CreateFormFile(name, filename)
	part.Write(content)
	return b
}

// Build finalizes the multipart body.
// Returns the Content-Type header value and body bytes, ready to pass to DoRaw.
// Must be called exactly once.
func (b *MultipartBuilder) Build() (contentType string, body []byte) {
	b.writer.Close()
	return b.writer.FormDataContentType(), b.buf.Bytes()
}
