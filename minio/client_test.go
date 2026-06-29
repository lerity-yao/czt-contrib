package minio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newFakeMinioServer creates a httptest server for minio-go client initialization.
// It wraps the given handler with bucket location handling required by minio-go.
func newFakeMinioServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// minio-go always queries bucket location before operations.
		if r.URL.Query().Get("location") != "" || r.URL.RawQuery == "location=" || strings.Contains(r.URL.RawQuery, "location") {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/"></LocationConstraint>`)
			return
		}
		if handler != nil {
			handler(w, r)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
}

// hostFromServer extracts host:port from the server URL.
func hostFromServer(server *httptest.Server) string {
	return server.Listener.Addr().String()
}

func validConf(endpoints ...string) Conf {
	return Conf{
		Endpoints:        endpoints,
		AccessKeyID:      "testkey",
		SecretAccessKey:  "testsecret",
		SignatureVersion: SignatureV4,
		SlowThreshold:    1000,
	}
}

// --- NewClient tests ---

func TestNewClient_Success(t *testing.T) {
	server := newFakeMinioServer(nil)
	defer server.Close()

	c, err := NewClient(validConf(hostFromServer(server)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_MultipleEndpoints(t *testing.T) {
	s1 := newFakeMinioServer(nil)
	defer s1.Close()
	s2 := newFakeMinioServer(nil)
	defer s2.Close()

	c, err := NewClient(validConf(hostFromServer(s1), hostFromServer(s2)))
	if err != nil {
		t.Fatal(err)
	}
	clients := c.RawClients()
	if len(clients) != 2 {
		t.Fatalf("expected 2 raw clients, got %d", len(clients))
	}
}

func TestNewClient_V2Signature(t *testing.T) {
	server := newFakeMinioServer(nil)
	defer server.Close()
	conf := validConf(hostFromServer(server))
	conf.SignatureVersion = SignatureV2

	c, err := NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		conf Conf
	}{
		{"empty config", Conf{}},
		{"no endpoints", Conf{
			AccessKeyID: "a", SecretAccessKey: "b", SignatureVersion: SignatureV4,
		}},
		{"empty access key", Conf{
			Endpoints: []string{"ep:9000"}, SecretAccessKey: "b", SignatureVersion: SignatureV4,
		}},
		{"empty secret key", Conf{
			Endpoints: []string{"ep:9000"}, AccessKeyID: "a", SignatureVersion: SignatureV4,
		}},
		{"bad signature version", Conf{
			Endpoints: []string{"ep:9000"}, AccessKeyID: "a", SecretAccessKey: "b", SignatureVersion: "v3",
		}},
		{"negative slow threshold", Conf{
			Endpoints: []string{"ep:9000"}, AccessKeyID: "a", SecretAccessKey: "b",
			SignatureVersion: SignatureV4, SlowThreshold: -1,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.conf)
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
		})
	}
}

func TestNewClient_WithTransport(t *testing.T) {
	server := newFakeMinioServer(nil)
	defer server.Close()
	custom := &http.Transport{}
	c, err := NewClient(validConf(hostFromServer(server)), WithTransport(custom))
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

// Note: MustNewClient calls logx.Must which calls os.Exit(1), not panic — untestable directly.
// We test the success path and verify NewClient returns error for invalid conf.
func TestMustNewClient_Success(t *testing.T) {
	server := newFakeMinioServer(nil)
	defer server.Close()
	c := MustNewClient(validConf(hostFromServer(server)))
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestMustNewClient_InvalidConf(t *testing.T) {
	// Verify the underlying NewClient returns an error for the same invalid conf.
	_, err := NewClient(Conf{})
	if err == nil {
		t.Fatal("expected error for invalid conf")
	}
}

// --- Options tests ---

func TestWithContentType(t *testing.T) {
	o := &requestOptions{}
	WithContentType("text/plain")(o)
	if o.contentType != "text/plain" {
		t.Fatalf("expected 'text/plain', got '%s'", o.contentType)
	}
}

func TestWithMetadata(t *testing.T) {
	o := &requestOptions{}
	m := map[string]string{"key": "value"}
	WithMetadata(m)(o)
	if o.metadata["key"] != "value" {
		t.Fatal("metadata not set")
	}
}

func TestWithStorageClass(t *testing.T) {
	o := &requestOptions{}
	WithStorageClass("STANDARD")(o)
	if o.storageClass != "STANDARD" {
		t.Fatal("storage class not set")
	}
}

func TestWithPartSize(t *testing.T) {
	o := &requestOptions{}
	WithPartSize(32 * 1024 * 1024)(o)
	if o.partSize != 32*1024*1024 {
		t.Fatal("part size not set")
	}
}

func TestBuildRequestOptions_Defaults(t *testing.T) {
	cc := &CommonClient{}
	o := cc.buildRequestOptions(nil)
	if o.contentType != defaultContentType {
		t.Fatalf("expected default content type, got '%s'", o.contentType)
	}
	if o.partSize != defaultPartSize {
		t.Fatalf("expected default part size, got %d", o.partSize)
	}
}

func TestBuildRequestOptions_WithAllOptions(t *testing.T) {
	cc := &CommonClient{}
	opts := []Option{
		WithContentType("image/png"),
		WithMetadata(map[string]string{"author": "test"}),
		WithStorageClass("REDUCED_REDUNDANCY"),
		WithPartSize(10 * 1024 * 1024),
	}
	o := cc.buildRequestOptions(opts)
	if o.contentType != "image/png" {
		t.Fatalf("expected 'image/png', got '%s'", o.contentType)
	}
	if o.metadata["author"] != "test" {
		t.Fatal("metadata not set correctly")
	}
	if o.storageClass != "REDUCED_REDUNDANCY" {
		t.Fatal("storage class not set correctly")
	}
	if o.partSize != 10*1024*1024 {
		t.Fatalf("expected part size 10MB, got %d", o.partSize)
	}
}

// --- RawClient / RawClients tests ---

func TestRawClient(t *testing.T) {
	server := newFakeMinioServer(nil)
	defer server.Close()
	c, _ := NewClient(validConf(hostFromServer(server)))
	raw := c.RawClient()
	if raw == nil {
		t.Fatal("expected non-nil raw client")
	}
}

func TestRawClients(t *testing.T) {
	s1 := newFakeMinioServer(nil)
	defer s1.Close()
	s2 := newFakeMinioServer(nil)
	defer s2.Close()
	c, _ := NewClient(validConf(hostFromServer(s1), hostFromServer(s2)))
	clients := c.RawClients()
	if len(clients) != 2 {
		t.Fatalf("expected 2, got %d", len(clients))
	}
	for i, cl := range clients {
		if cl == nil {
			t.Fatalf("client at index %d is nil", i)
		}
	}
}

// --- Convenience method tests ---

func TestDelete_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	err := c.Delete(context.Background(), "testbucket", "testkey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExists_True(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("ETag", `"abc123"`)
		w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	exists, err := c.Exists(context.Background(), "testbucket", "testkey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatal("expected exists=true")
	}
}

func TestExists_False_NoSuchKey(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchKey</Code><Message>The specified key does not exist.</Message></Error>`)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	exists, err := c.Exists(context.Background(), "testbucket", "nokey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatal("expected exists=false")
	}
}

func TestGetPresignedDownloadURL(t *testing.T) {
	server := newFakeMinioServer(nil)
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	url, err := c.GetPresignedDownloadURL(context.Background(), "bucket", "key", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	if !strings.Contains(url, "bucket") {
		t.Fatal("URL should contain bucket name")
	}
}

func TestGetPresignedUploadURL(t *testing.T) {
	server := newFakeMinioServer(nil)
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	url, err := c.GetPresignedUploadURL(context.Background(), "bucket", "key", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
}

func TestUploadReader_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			io.Copy(io.Discard, r.Body)
			w.Header().Set("ETag", `"etag123"`)
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	reader := strings.NewReader("hello world")
	info, err := c.UploadReader(context.Background(), "bucket", "key", reader, 11,
		WithContentType("text/plain"), WithMetadata(map[string]string{"x": "y"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil upload info")
	}
	if info.Key != "key" {
		t.Fatalf("expected key='key', got '%s'", info.Key)
	}
}

func TestDownload_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "5")
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("ETag", `"abc"`)
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	rc, err := c.Download(context.Background(), "bucket", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rc.Close()
	data, _ := io.ReadAll(rc)
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", string(data))
	}
}

func TestDownload_NotFound(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchKey</Code><Message>not found</Message></Error>`)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	_, err := c.Download(context.Background(), "bucket", "nokey")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestDownload_NoEndpoints(t *testing.T) {
	cc := newTestCommonClient(nil)
	_, err := cc.Download(context.Background(), "bucket", "key")
	if err == nil {
		t.Fatal("expected error for no endpoints")
	}
}

func TestUploadFile_Error(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><Error><Code>AccessDenied</Code><Message>Denied</Message></Error>`)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	_, err := c.UploadFile(context.Background(), "bucket", "key", "/nonexistent/file.txt")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Failover integration test ---

func TestDelete_Failover(t *testing.T) {
	s1 := newFakeMinioServer(nil)
	addr1 := hostFromServer(s1)
	s1.Close() // closed = network error

	s2 := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer s2.Close()

	c, err := NewClient(validConf(addr1, hostFromServer(s2)))
	if err != nil {
		t.Fatal(err)
	}
	err = c.Delete(context.Background(), "bucket", "key")
	if err != nil {
		t.Fatalf("expected failover success, got: %v", err)
	}
}
