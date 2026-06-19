package gateway

import (
	"bytes"
	"mime/multipart"
	"testing"
)

func TestMultipartBuilder(t *testing.T) {
	ct, body, err := NewMultipart().
		Field("name", "tom").
		File("avatar", "avatar.png", []byte("png-bytes")).
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if ct == "" {
		t.Error("Content-Type should not be empty")
	}
	if !bytes.Contains(body, []byte(`name="name"`)) || !bytes.Contains(body, []byte("tom")) {
		t.Error("body should contain text field 'name=tom'")
	}
	if !bytes.Contains(body, []byte(`name="avatar"`)) || !bytes.Contains(body, []byte("avatar.png")) {
		t.Error("body should contain file field 'avatar'")
	}
}

func TestMultipartBuilder_Empty(t *testing.T) {
	ct, body, err := NewMultipart().Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if ct == "" {
		t.Error("Content-Type should not be empty")
	}
	if len(body) == 0 {
		t.Error("body should not be empty for empty multipart")
	}
}

func TestMultipartBuilder_ErrorPropagates(t *testing.T) {
	// Replace the underlying writer with one that always fails.
	b := NewMultipart()
	b.writer = multipart.NewWriter(&failingWriter{})
	_ = b.Field("name", "tom")
	if b.err == nil {
		t.Fatal("expected error from failing writer")
	}
	_, _, err := b.Build()
	if err == nil {
		t.Fatal("Build() should return accumulated error")
	}
}

type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (int, error) {
	return 0, multipart.ErrMessageTooLarge
}
