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
	err    error
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
	if b.err != nil {
		return b
	}
	if err := b.writer.WriteField(name, value); err != nil {
		b.err = err
	}
	return b
}

// File adds a file field to the multipart form.
// Internally uses multipart.Writer.CreateFormFile.
func (b *MultipartBuilder) File(name, filename string, content []byte) *MultipartBuilder {
	if b.err != nil {
		return b
	}
	part, err := b.writer.CreateFormFile(name, filename)
	if err != nil {
		b.err = err
		return b
	}
	if _, err := part.Write(content); err != nil {
		b.err = err
	}
	return b
}

// Build finalizes the multipart body.
// Returns the Content-Type header value, body bytes, and any error encountered
// during Field/File operations or writer.Close().
// Must be called exactly once.
func (b *MultipartBuilder) Build() (contentType string, body []byte, err error) {
	if b.err != nil {
		return "", nil, b.err
	}
	if err := b.writer.Close(); err != nil {
		return "", nil, err
	}
	return b.writer.FormDataContentType(), b.buf.Bytes(), nil
}
