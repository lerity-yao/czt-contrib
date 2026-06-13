package gateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// --------------- sortedQuery ---------------

func TestSortedQuery(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"empty", "", ""},
		{"single", "b=2", "?b=2"},
		{"already sorted", "a=1&b=2", "?a=1&b=2"},
		{"needs sort", "b=2&a=1", "?a=1&b=2"},
		{"duplicate keys first wins", "a=1&a=2", "?a=1"},
		{"empty value omits equals", "a=", "?a"},
		{"no value no equals", "a", "?a"},
		{"mixed", "c=3&a&b=2&a=1", "?a&b=2&c=3"},
		{"url-encoded preserved", "name=%E5%BC%A0&age=20", "?age=20&name=%E5%BC%A0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortedQuery(tt.input)
			if got != tt.expect {
				t.Errorf("sortedQuery(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

// --------------- splitQueryPart ---------------

func TestSplitQueryPart(t *testing.T) {
	tests := []struct {
		input    string
		wantKey  string
		wantSign string
	}{
		{"key=value", "key", "key=value"},
		{"key=", "key", "key"},
		{"key", "key", "key"},
		{"a=hello", "a", "a=hello"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			key, sign := splitQueryPart(tt.input)
			if key != tt.wantKey || sign != tt.wantSign {
				t.Errorf("splitQueryPart(%q) = (%q, %q), want (%q, %q)",
					tt.input, key, sign, tt.wantKey, tt.wantSign)
			}
		})
	}
}

// --------------- md5Hash ---------------

func TestMd5Hash(t *testing.T) {
	// MD5("") base64
	got := md5Hash([]byte(""))
	if got != "1B2M2Y8AsgTpgAmY7PhCfg==" {
		t.Errorf("md5Hash(empty) = %q, want 1B2M2Y8AsgTpgAmY7PhCfg==", got)
	}

	// MD5("hello") base64
	got = md5Hash([]byte("hello"))
	if got != "XUFAKrxLKna5cZ2REBfFkg==" {
		t.Errorf("md5Hash(hello) = %q, want XUFAKrxLKna5cZ2REBfFkg==", got)
	}
}

// --------------- hmacSHA256 ---------------

func TestHmacSHA256(t *testing.T) {
	data := "POST\n*/*\n\napplication/json\n"
	key := "test-secret"

	got := hmacSHA256(data, []byte(key))

	// Independently compute expected value
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	expected := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if got != expected {
		t.Errorf("hmacSHA256 mismatch: got %q, want %q", got, expected)
	}
}

// --------------- buildStringToSign ---------------

func TestBuildStringToSign(t *testing.T) {
	r := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Path:     "/v1/users",
			RawQuery: "name=tom&page=1",
		},
		Header: http.Header{},
	}
	r.Header.Set(headerAccept, "*/*")
	r.Header.Set(headerContentMD5, "abc123==")
	r.Header.Set(headerContentType, "application/json")
	r.Header.Set(headerDate, "Wed, 13 Jun 2026 12:00:00 GMT")
	r.Header.Set(headerCaKey, "myappkey")
	r.Header.Set(headerCaNonce, "nonce-123")
	r.Header.Set(headerCaSignatureMethod, "HmacSHA256")
	r.Header.Set(headerCaTimestamp, "1718280000000")

	sts := buildStringToSign(r)

	expected := strings.Join([]string{
		"POST",
		"*/*",
		"abc123==",
		"application/json",
		"Wed, 13 Jun 2026 12:00:00 GMT",
		"x-ca-key:myappkey",
		"x-ca-nonce:nonce-123",
		"x-ca-signature-method:HmacSHA256",
		"x-ca-timestamp:1718280000000",
	}, "\n") + "\n" + "/v1/users" + "?name=tom&page=1"

	if sts != expected {
		t.Errorf("buildStringToSign mismatch:\ngot:  %q\nwant: %q", sts, expected)
	}
}

func TestBuildStringToSign_NoQuery(t *testing.T) {
	r := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/v1/users"},
		Header: http.Header{},
	}
	r.Header.Set(headerAccept, "*/*")

	sts := buildStringToSign(r)

	// No query → URL part is just the path
	if !strings.HasSuffix(sts, "/v1/users") {
		t.Errorf("should end with path, got: %q", sts)
	}
	// No Content-MD5, Content-Type, Date → empty lines
	if !strings.Contains(sts, "GET\n*/*\n\n\n") {
		t.Errorf("empty header lines not found in: %q", sts)
	}
}

func TestBuildStringToSign_QuerySorted(t *testing.T) {
	r := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path:     "/api",
			RawQuery: "z=3&a=1&m=2",
		},
		Header: http.Header{},
	}

	sts := buildStringToSign(r)

	// Query should be sorted in the signature string
	if !strings.HasSuffix(sts, "/api?a=1&m=2&z=3") {
		t.Errorf("query should be sorted, got: %q", sts)
	}
}

// --------------- signRequest ---------------

func TestSignRequest_DefaultHeaders(t *testing.T) {
	r := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/test"},
		Header: http.Header{},
	}

	signRequest(r, "appkey", []byte("secret"), nil)

	if r.Header.Get(headerAccept) != defaultAccept {
		t.Errorf("Accept = %q, want %q", r.Header.Get(headerAccept), defaultAccept)
	}
	if r.Header.Get(headerUserAgent) != defaultUserAgent {
		t.Errorf("User-Agent = %q, want %q", r.Header.Get(headerUserAgent), defaultUserAgent)
	}
	if r.Header.Get(headerDate) == "" {
		t.Error("Date should be set")
	}
}

func TestSignRequest_PreservesExistingHeaders(t *testing.T) {
	r := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/test"},
		Header: http.Header{},
	}
	r.Header.Set(headerAccept, "application/json")
	r.Header.Set(headerDate, "Custom Date")

	signRequest(r, "appkey", []byte("secret"), nil)

	if r.Header.Get(headerAccept) != "application/json" {
		t.Errorf("Accept should be preserved, got %q", r.Header.Get(headerAccept))
	}
	if r.Header.Get(headerDate) != "Custom Date" {
		t.Errorf("Date should be preserved, got %q", r.Header.Get(headerDate))
	}
}

func TestSignRequest_ContentMD5(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        []byte
		wantMD5     bool
	}{
		{"json body", "application/json", []byte(`{"name":"tom"}`), true},
		{"xml body", "application/xml", []byte(`<xml/>`), true},
		{"empty body", "application/json", nil, false},
		{"form body", "application/x-www-form-urlencoded", []byte("a=1"), false},
		{"form with charset", "application/x-www-form-urlencoded; charset=utf-8", []byte("a=1"), false},
		{"multipart body", "multipart/form-data; boundary=xxx", []byte("xxx"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{
				Method: "POST",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{},
			}
			if tt.contentType != "" {
				r.Header.Set(headerContentType, tt.contentType)
			}

			signRequest(r, "appkey", []byte("secret"), tt.body)

			hasMD5 := r.Header.Get(headerContentMD5) != ""
			if hasMD5 != tt.wantMD5 {
				t.Errorf("Content-MD5 present = %v, want %v (value: %q)",
					hasMD5, tt.wantMD5, r.Header.Get(headerContentMD5))
			}
		})
	}
}

func TestSignRequest_CaHeaders(t *testing.T) {
	r := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/test"},
		Header: http.Header{},
	}

	signRequest(r, "myappkey", []byte("mysecret"), nil)

	if r.Header.Get(headerCaKey) != "myappkey" {
		t.Errorf("X-Ca-Key = %q, want myappkey", r.Header.Get(headerCaKey))
	}
	if r.Header.Get(headerCaNonce) == "" {
		t.Error("X-Ca-Nonce should not be empty")
	}
	if r.Header.Get(headerCaSignatureMethod) != defaultSignMethod {
		t.Errorf("X-Ca-Signature-Method = %q, want %q", r.Header.Get(headerCaSignatureMethod), defaultSignMethod)
	}
	if r.Header.Get(headerCaTimestamp) == "" {
		t.Error("X-Ca-Timestamp should not be empty")
	}
	if r.Header.Get(headerCaSignature) == "" {
		t.Error("X-Ca-Signature should not be empty")
	}
}

func TestSignRequest_SignatureHeadersValue(t *testing.T) {
	r := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/test"},
		Header: http.Header{},
	}

	signRequest(r, "key", []byte("secret"), nil)

	expected := "x-ca-key,x-ca-nonce,x-ca-signature-method,x-ca-timestamp"
	if r.Header.Get(headerCaSignatureHeaders) != expected {
		t.Errorf("X-Ca-Signature-Headers = %q, want %q",
			r.Header.Get(headerCaSignatureHeaders), expected)
	}
}

func TestSignRequest_SignatureVerifiable(t *testing.T) {
	r := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Path:     "/v1/users",
			RawQuery: "page=1",
		},
		Header: http.Header{},
	}
	r.Header.Set(headerContentType, "application/json")

	body := []byte(`{"name":"tom"}`)
	appKey := "test-key"
	appSecret := "test-secret"

	signRequest(r, appKey, []byte(appSecret), body)

	// Recompute signature from the request's final header state
	sts := buildStringToSign(r)
	expectedSig := hmacSHA256(sts, []byte(appSecret))

	if r.Header.Get(headerCaSignature) != expectedSig {
		t.Errorf("Signature mismatch:\nrequest:    %q\nrecomputed: %q",
			r.Header.Get(headerCaSignature), expectedSig)
	}
}

func TestSignRequest_NonceUnique(t *testing.T) {
	// Each call to signRequest should produce a unique nonce
	r1 := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/test"},
		Header: http.Header{},
	}
	r2 := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/test"},
		Header: http.Header{},
	}

	signRequest(r1, "key", []byte("secret"), nil)
	signRequest(r2, "key", []byte("secret"), nil)

	n1 := r1.Header.Get(headerCaNonce)
	n2 := r2.Header.Get(headerCaNonce)

	if n1 == n2 {
		t.Errorf("Nonce should be unique, both were %q", n1)
	}
}
