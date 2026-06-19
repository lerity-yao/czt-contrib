package gateway

import (
	"context"
	"errors"
	"net/http"
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
