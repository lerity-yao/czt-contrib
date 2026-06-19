package gateway

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient_ValidationError(t *testing.T) {
	_, err := NewClient(Conf{})
	if err == nil {
		t.Fatal("expected error for empty config")
	}
}

func TestNewClient_DefaultHTTPClient(t *testing.T) {
	c, err := NewClient(Conf{
		Host:      "https://api.example.com",
		AppKey:    "key",
		AppSecret: "secret",
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	cc := c.(*CommonClient)
	if cc.cli != http.DefaultClient {
		t.Error("default client should be http.DefaultClient")
	}
}

func TestNewClient_WithClient(t *testing.T) {
	custom := &http.Client{}
	c, err := NewClient(Conf{
		Host:      "https://api.example.com",
		AppKey:    "key",
		AppSecret: "secret",
	}, WithClient(custom))
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	cc := c.(*CommonClient)
	if cc.cli != custom {
		t.Error("custom client not injected")
	}
}

func TestMustNewClient(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustNewClient panicked: %v", r)
		}
	}()

	_ = MustNewClient(Conf{
		Host:      "https://api.example.com",
		AppKey:    "key",
		AppSecret: "secret",
	})
}

func TestDoRaw_MissingContentType(t *testing.T) {
	c, err := NewClient(Conf{
		Host:      "https://api.example.com",
		AppKey:    "key",
		AppSecret: "secret",
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	_, err = c.DoRaw(context.Background(), http.MethodPost, "/test", "", []byte("body"))
	if err == nil {
		t.Fatal("expected error when body is present but contentType is empty")
	}
}

func TestParse_WithError(t *testing.T) {
	err := errors.New("upstream error")
	if got := Parse(nil, err, nil); got != err {
		t.Errorf("Parse should return upstream error, got %v", got)
	}
}

func TestParse_Success(t *testing.T) {
	body := `{"id":123}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := Parse(resp, nil, &result); err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if result.ID != 123 {
		t.Errorf("ID = %d, want 123", result.ID)
	}
}

func TestDoRaw_JSONBody(t *testing.T) {
	var receivedSig, receivedKey, receivedMD5 string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSig = r.Header.Get(headerCaSignature)
		receivedKey = r.Header.Get(headerCaKey)
		receivedMD5 = r.Header.Get(headerContentMD5)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client, err := NewClient(Conf{
		Host:      server.URL,
		AppKey:    "test-key",
		AppSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	body := []byte(`{"name":"tom"}`)
	resp, err := client.DoRaw(context.Background(), http.MethodPost, "/test", "application/json", body)
	if err != nil {
		t.Fatalf("DoRaw error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if receivedSig == "" {
		t.Error("X-Ca-Signature not received")
	}
	if receivedKey != "test-key" {
		t.Errorf("X-Ca-Key = %q, want test-key", receivedKey)
	}
	if receivedMD5 == "" {
		t.Error("Content-MD5 not received for JSON body")
	}
}

func TestDoRaw_FormBody(t *testing.T) {
	var receivedMD5 string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMD5 = r.Header.Get(headerContentMD5)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(Conf{
		Host:      server.URL,
		AppKey:    "test-key",
		AppSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	body := []byte("a=1&b=2")
	resp, err := client.DoRaw(context.Background(), http.MethodPost, "/test", contentTypeForm, body)
	if err != nil {
		t.Fatalf("DoRaw error: %v", err)
	}
	defer resp.Body.Close()

	if receivedMD5 != "" {
		t.Error("Content-MD5 should not be set for form body")
	}
}

func TestDoRaw_NoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(headerCaSignature) == "" {
			t.Error("X-Ca-Signature not received")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(Conf{
		Host:      server.URL,
		AppKey:    "test-key",
		AppSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	resp, err := client.DoRaw(context.Background(), http.MethodGet, "/test", "", nil)
	if err != nil {
		t.Fatalf("DoRaw error: %v", err)
	}
	defer resp.Body.Close()
}

func TestDo_GET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/test" {
			t.Errorf("path = %q, want /test", r.URL.Path)
		}
		if r.Header.Get(headerCaSignature) == "" {
			t.Error("X-Ca-Signature not received")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client, err := NewClient(Conf{
		Host:      server.URL,
		AppKey:    "test-key",
		AppSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	resp, err := client.Do(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("Do error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestDo_POST_JSON(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		if r.Header.Get(headerCaSignature) == "" {
			t.Error("X-Ca-Signature not received")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(Conf{
		Host:      server.URL,
		AppKey:    "test-key",
		AppSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	type reqData struct {
		Name string `json:"name"`
	}
	resp, err := client.Do(context.Background(), http.MethodPost, "/test", &reqData{Name: "tom"})
	if err != nil {
		t.Fatalf("Do error: %v", err)
	}
	defer resp.Body.Close()

	if !bytes.Contains(receivedBody, []byte(`"name":"tom"`)) {
		t.Errorf("body = %q, want containing name:tom", receivedBody)
	}
}

func TestSignOption_BufferBody(t *testing.T) {
	c := Conf{AppKey: "key", AppSecret: "secret"}
	opt := signOption(c)

	body := []byte(`{"name":"tom"}`)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://api.example.com/test", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest error: %v", err)
	}
	req.Header.Set(headerContentType, "application/json")

	opt(req)

	if req.Header.Get(headerContentMD5) == "" {
		t.Error("Content-MD5 should be set for JSON body")
	}
	if req.Header.Get(headerCaSignature) == "" {
		t.Error("X-Ca-Signature should be set")
	}

	// Body should still be readable after signing.
	gotBody, _ := io.ReadAll(req.Body)
	if !bytes.Equal(gotBody, body) {
		t.Errorf("body after signing = %q, want %q", gotBody, body)
	}
}

func TestSignOption_SkipBufferForFormWithGetBody(t *testing.T) {
	c := Conf{AppKey: "key", AppSecret: "secret"}
	opt := signOption(c)

	body := []byte("a=1")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://api.example.com/test", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest error: %v", err)
	}
	req.Header.Set(headerContentType, contentTypeForm)

	opt(req)

	if req.Header.Get(headerContentMD5) != "" {
		t.Error("Content-MD5 should not be set for form body")
	}
	if req.Header.Get(headerCaSignature) == "" {
		t.Error("X-Ca-Signature should be set")
	}
}

func TestSignOption_SkipBufferForFormWithoutGetBody(t *testing.T) {
	c := Conf{AppKey: "key", AppSecret: "secret"}
	opt := signOption(c)

	body := []byte("a=1")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://api.example.com/test", io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		t.Fatalf("NewRequest error: %v", err)
	}
	req.Header.Set(headerContentType, contentTypeForm)
	req.GetBody = nil

	opt(req)

	if req.Header.Get(headerCaSignature) == "" {
		t.Error("X-Ca-Signature should be set")
	}
}
