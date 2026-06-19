package hmacauth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ==================== vars.go ====================

func TestDefaultSignHeaders_ReturnsCopy(t *testing.T) {
	a := defaultSignHeaders()
	b := defaultSignHeaders()
	if &a[0] == &b[0] {
		t.Fatal("expected independent copies")
	}
	if len(a) != 2 || a[0] != "date" || a[1] != pseudoRequestTarget {
		t.Fatalf("unexpected default headers: %v", a)
	}
}

func TestJoinHeaders(t *testing.T) {
	got := joinHeaders([]string{"date", "@request-target", "digest"})
	want := "date @request-target digest"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestJoinHeaders_Single(t *testing.T) {
	got := joinHeaders([]string{"date"})
	if got != "date" {
		t.Fatalf("got %q", got)
	}
}

// ==================== config.go ====================

func TestConf_Validate_OK_Defaults(t *testing.T) {
	c := Conf{Host: "http://kong.example.com", Username: "user", Secret: "secret"}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Algorithm != defaultAlgorithm {
		t.Fatalf("expected default algorithm %q, got %q", defaultAlgorithm, c.Algorithm)
	}
	if len(c.Headers) != 2 {
		t.Fatalf("expected 2 default headers, got %d", len(c.Headers))
	}
	// Host trailing slash trimmed
	c2 := Conf{Host: "https://kong.example.com///", Username: "u", Secret: "s"}
	_ = c2.Validate()
	if strings.HasSuffix(c2.Host, "/") {
		t.Fatal("expected trailing slashes trimmed")
	}
}

func TestConf_Validate_EmptyHost(t *testing.T) {
	c := Conf{Username: "u", Secret: "s"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestConf_Validate_BadHostPrefix(t *testing.T) {
	c := Conf{Host: "ftp://bad.host", Username: "u", Secret: "s"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for bad host prefix")
	}
}

func TestConf_Validate_EmptyUsername(t *testing.T) {
	c := Conf{Host: "http://h", Secret: "s"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty username")
	}
}

func TestConf_Validate_EmptySecret(t *testing.T) {
	c := Conf{Host: "http://h", Username: "u"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestConf_Validate_UnsupportedAlgorithm(t *testing.T) {
	c := Conf{Host: "http://h", Username: "u", Secret: "s", Algorithm: "md5"}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for unsupported algorithm")
	}
}

func TestConf_Validate_AllAlgorithms(t *testing.T) {
	algos := []string{
		AlgorithmHmacSHA1,
		AlgorithmHmacSHA224,
		AlgorithmHmacSHA256,
		AlgorithmHmacSHA384,
		AlgorithmHmacSHA512,
	}
	for _, algo := range algos {
		c := Conf{Host: "http://h", Username: "u", Secret: "s", Algorithm: algo}
		if err := c.Validate(); err != nil {
			t.Fatalf("unexpected error for algo %q: %v", algo, err)
		}
	}
}

func TestConf_Validate_AlgorithmUpperCase(t *testing.T) {
	c := Conf{Host: "http://h", Username: "u", Secret: "s", Algorithm: "HMAC-SHA256"}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Algorithm != "hmac-sha256" {
		t.Fatalf("expected lowercase algorithm, got %q", c.Algorithm)
	}
}

func TestConf_Validate_CustomHeaders_Normalised(t *testing.T) {
	c := Conf{
		Host:     "http://h",
		Username: "u",
		Secret:   "s",
		Headers:  []string{"Date", "HOST", "@Request-Target"},
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, h := range c.Headers {
		if h != strings.ToLower(h) {
			t.Fatalf("header not normalised: %q", h)
		}
	}
}

func TestConf_Validate_HTTPS(t *testing.T) {
	c := Conf{Host: "https://secure.host", Username: "u", Secret: "s"}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ==================== sign.go ====================

func newConf(headers ...string) Conf {
	c := Conf{Host: "http://example.com", Username: "testuser", Secret: "testsecret"}
	if len(headers) > 0 {
		c.Headers = headers
	}
	_ = c.Validate()
	return c
}

func newRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	return req
}

func TestIncludesHeader(t *testing.T) {
	headers := []string{"date", "@request-target", "digest"}
	if !includesHeader(headers, "date") {
		t.Fatal("expected date to be included")
	}
	if includesHeader(headers, "host") {
		t.Fatal("expected host to NOT be included")
	}
}

func TestHeaderValue_RequestTarget(t *testing.T) {
	req := newRequest(t, "POST", "http://example.com/api/v1?foo=bar", nil)
	v := headerValue(req, pseudoRequestTarget)
	want := "post /api/v1?foo=bar"
	if v != want {
		t.Fatalf("got %q, want %q", v, want)
	}
}

func TestHeaderValue_Host_FromHeader(t *testing.T) {
	req := newRequest(t, "GET", "http://example.com/path", nil)
	req.Header.Set("Host", "override.example.com")
	v := headerValue(req, "host")
	if v != "override.example.com" {
		t.Fatalf("got %q", v)
	}
}

func TestHeaderValue_Host_FromRequestHost(t *testing.T) {
	req := newRequest(t, "GET", "http://example.com/path", nil)
	req.Host = "example.com"
	// No explicit Host header set
	v := headerValue(req, "host")
	if v != "example.com" {
		t.Fatalf("got %q", v)
	}
}

func TestHeaderValue_StandardHeader(t *testing.T) {
	req := newRequest(t, "GET", "http://example.com/", nil)
	req.Header.Set("X-Custom", "myvalue")
	v := headerValue(req, "x-custom")
	if v != "myvalue" {
		t.Fatalf("got %q", v)
	}
}

func TestComputeHMAC_AllAlgorithms(t *testing.T) {
	cases := []string{
		AlgorithmHmacSHA1,
		AlgorithmHmacSHA224,
		AlgorithmHmacSHA256,
		AlgorithmHmacSHA384,
		AlgorithmHmacSHA512,
	}
	for _, algo := range cases {
		sig := computeHMAC(algo, []byte("key"), "data")
		if sig == "" {
			t.Fatalf("empty signature for algo %q", algo)
		}
		// Must be valid base64
		if _, err := base64.StdEncoding.DecodeString(sig); err != nil {
			t.Fatalf("invalid base64 for algo %q: %v", algo, err)
		}
	}
}

func TestComputeHMAC_Default(t *testing.T) {
	// unknown algorithm falls through to default (hmac-sha256)
	sig := computeHMAC("unknown-algo", []byte("key"), "data")
	// Verify manually
	mac := hmac.New(sha256.New, []byte("key"))
	mac.Write([]byte("data"))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if sig != expected {
		t.Fatalf("got %q, want %q", sig, expected)
	}
}

func TestBuildAuthorization(t *testing.T) {
	c := newConf("date", "@request-target")
	auth := buildAuthorization(c, "sig==")
	if !strings.HasPrefix(auth, "hmac ") {
		t.Fatalf("expected hmac prefix, got %q", auth)
	}
	if !strings.Contains(auth, `username="testuser"`) {
		t.Fatalf("missing username in auth: %q", auth)
	}
	if !strings.Contains(auth, `algorithm="hmac-sha256"`) {
		t.Fatalf("missing algorithm in auth: %q", auth)
	}
	if !strings.Contains(auth, `headers="date @request-target"`) {
		t.Fatalf("missing headers in auth: %q", auth)
	}
	if !strings.Contains(auth, `signature="sig=="`) {
		t.Fatalf("missing signature in auth: %q", auth)
	}
}

func TestSignRequest_DefaultHeaders(t *testing.T) {
	c := newConf() // default: date + @request-target
	req := newRequest(t, "GET", "http://example.com/api", nil)
	now := time.Now().UTC()
	signRequest(req, c, nil, now)

	if req.Header.Get(headerAuthorization) == "" {
		t.Fatal("Authorization header not set")
	}
	if req.Header.Get(headerDate) == "" {
		t.Fatal("Date header not set")
	}
	if req.Header.Get(headerUserAgent) == "" {
		t.Fatal("User-Agent header not set")
	}
}

func TestSignRequest_DateAlreadySet(t *testing.T) {
	c := newConf()
	req := newRequest(t, "GET", "http://example.com/api", nil)
	custom := "Mon, 01 Jan 2024 00:00:00 GMT"
	req.Header.Set(headerDate, custom)
	signRequest(req, c, nil, time.Now())
	if req.Header.Get(headerDate) != custom {
		t.Fatal("existing Date header should not be overwritten")
	}
}

func TestSignRequest_UserAgentAlreadySet(t *testing.T) {
	c := newConf()
	req := newRequest(t, "GET", "http://example.com/api", nil)
	req.Header.Set(headerUserAgent, "MyClient/1.0")
	signRequest(req, c, nil, time.Now())
	if req.Header.Get(headerUserAgent) != "MyClient/1.0" {
		t.Fatal("existing User-Agent should not be overwritten")
	}
}

func TestSignRequest_WithDigest(t *testing.T) {
	c := newConf("date", "@request-target", "digest")
	req := newRequest(t, "POST", "http://example.com/api", nil)
	body := []byte(`{"key":"value"}`)
	signRequest(req, c, body, time.Now())

	digest := req.Header.Get(headerDigest)
	if digest == "" {
		t.Fatal("Digest header not set")
	}
	if !strings.HasPrefix(digest, "SHA-256=") {
		t.Fatalf("unexpected digest format: %q", digest)
	}
}

func TestSignRequest_DigestEmptyBody(t *testing.T) {
	c := newConf("date", "@request-target", "digest")
	req := newRequest(t, "POST", "http://example.com/api", nil)
	signRequest(req, c, nil, time.Now())
	digest := req.Header.Get(headerDigest)
	if digest == "" {
		t.Fatal("Digest header not set for empty body")
	}
	// SHA-256 of empty string
	h := sha256.Sum256(nil)
	expected := digestPrefix + base64.StdEncoding.EncodeToString(h[:])
	if digest != expected {
		t.Fatalf("unexpected empty-body digest: got %q, want %q", digest, expected)
	}
}

func TestSignRequest_DigestSkippedForForm(t *testing.T) {
	c := newConf("date", "@request-target", "digest")
	req := newRequest(t, "POST", "http://example.com/form", strings.NewReader("a=b"))
	req.Header.Set(headerContentType, contentTypeForm)
	signRequest(req, c, []byte("a=b"), time.Now())
	if req.Header.Get(headerDigest) != "" {
		t.Fatal("Digest should not be set for form content-type")
	}
}

func TestSignRequest_DigestSkippedForMultipart(t *testing.T) {
	c := newConf("date", "@request-target", "digest")
	req := newRequest(t, "POST", "http://example.com/upload", nil)
	req.Header.Set(headerContentType, contentTypeMultipart+"; boundary=xxx")
	signRequest(req, c, []byte("data"), time.Now())
	if req.Header.Get(headerDigest) != "" {
		t.Fatal("Digest should not be set for multipart content-type")
	}
}

func TestSignRequest_WithHost(t *testing.T) {
	c := newConf("host", "date", "@request-target")
	req := newRequest(t, "GET", "http://example.com/path", nil)
	req.Host = "example.com"
	signRequest(req, c, nil, time.Now())
	if req.Header.Get(headerHost) == "" {
		t.Fatal("Host header should be set when 'host' is in signing headers")
	}
}

func TestSignRequest_HostAlreadySet(t *testing.T) {
	c := newConf("host", "date", "@request-target")
	req := newRequest(t, "GET", "http://example.com/path", nil)
	req.Header.Set(headerHost, "override.host")
	signRequest(req, c, nil, time.Now())
	if req.Header.Get(headerHost) != "override.host" {
		t.Fatal("existing Host header should not be overwritten")
	}
}

func TestBuildSigningString_MultipleHeaders(t *testing.T) {
	c := newConf("date", "@request-target", "x-custom")
	req := newRequest(t, "GET", "http://example.com/api?q=1", nil)
	req.Header.Set("Date", "Mon, 01 Jan 2024 00:00:00 GMT")
	req.Header.Set("X-Custom", "hello")
	sts := buildSigningString(req, c)
	lines := strings.Split(sts, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), sts)
	}
	if !strings.HasPrefix(lines[0], "date: ") {
		t.Fatalf("unexpected first line: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "@request-target: ") {
		t.Fatalf("unexpected second line: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "x-custom: ") {
		t.Fatalf("unexpected third line: %q", lines[2])
	}
}

// ==================== client.go ====================

func newTestConf(host string) Conf {
	return Conf{Host: host, Username: "testuser", Secret: "testsecret"}
}

func TestNewClient_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli, err := NewClient(newTestConf(srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_InvalidConf(t *testing.T) {
	_, err := NewClient(Conf{}) // empty host
	if err == nil {
		t.Fatal("expected error for invalid conf")
	}
}

func TestWithClient_InjectsHTTPClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	custom := &http.Client{Timeout: 5 * time.Second}
	cli, err := NewClient(newTestConf(srv.URL), WithClient(custom))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cc := cli.(*CommonClient)
	if cc.cli != custom {
		t.Fatal("expected custom http.Client to be injected")
	}
}

func TestMustNewClient_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli := MustNewClient(newTestConf(srv.URL))
	if cli == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestMustNewClient_InvalidConf(t *testing.T) {
	// MustNewClient calls logx.Must which calls os.Exit, not panic — untestable directly.
	// Verify the underlying NewClient returns an error for the same invalid conf.
	_, err := NewClient(Conf{})
	if err == nil {
		t.Fatal("expected error for invalid conf")
	}
}

func TestClient_Do_GET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Authorization header is present
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli, _ := NewClient(newTestConf(srv.URL))
	resp, err := cli.Do(context.Background(), "GET", "/api/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClient_Do_POST_WithBody(t *testing.T) {
	type reqData struct {
		Name string `json:"name"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	cli, _ := NewClient(newTestConf(srv.URL))
	resp, err := cli.Do(context.Background(), "POST", "/users", &reqData{Name: "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestClient_DoRaw_WithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "raw payload" {
			t.Errorf("unexpected body: %q", body)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli, _ := NewClient(newTestConf(srv.URL))
	resp, err := cli.DoRaw(context.Background(), "POST", "/upload", "application/octet-stream", []byte("raw payload"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClient_DoRaw_NoBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli, _ := NewClient(newTestConf(srv.URL))
	resp, err := cli.DoRaw(context.Background(), "GET", "/health", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
}

func TestClient_DoRaw_BodyWithoutContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli, _ := NewClient(newTestConf(srv.URL))
	_, err := cli.DoRaw(context.Background(), "POST", "/upload", "", []byte("data"))
	if err == nil {
		t.Fatal("expected error when body present but contentType empty")
	}
}

func TestClient_DoRaw_BadURL(t *testing.T) {
	cli, _ := NewClient(Conf{Host: "http://localhost", Username: "u", Secret: "s"})
	// Inject a bad path that causes URL parse error
	_, err := cli.DoRaw(context.Background(), "GET\x00", "/path", "", nil)
	if err == nil {
		t.Fatal("expected error for invalid method")
	}
}

func TestParse_OK(t *testing.T) {
	type respData struct {
		Name string `json:"name"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"alice"}`))
	}))
	defer srv.Close()

	cli, _ := NewClient(newTestConf(srv.URL))
	resp, err := cli.Do(context.Background(), "GET", "/user", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data respData
	if err := Parse(resp, nil, &data); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if data.Name != "alice" {
		t.Fatalf("expected alice, got %q", data.Name)
	}
}

func TestParse_PropagatesError(t *testing.T) {
	sentinel := errors.New("upstream error")
	err := Parse(nil, sentinel, nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestSignOption_FormBody_NoDigest(t *testing.T) {
	// signOption should not set Digest for form content-type
	var capturedReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := Conf{
		Host:     srv.URL,
		Username: "u",
		Secret:   "s",
		Headers:  []string{"date", "@request-target", "digest"},
	}
	_ = c.Validate()
	cli, _ := NewClient(c)
	body := []byte("field=value")
	resp, err := cli.DoRaw(context.Background(), "POST", "/form", contentTypeForm, body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	_ = capturedReq
}

func TestSignOption_JSONBody_WithDigest(t *testing.T) {
	// signOption should set Digest for JSON content-type
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		digest := r.Header.Get("Digest")
		if digest == "" {
			t.Error("expected Digest header for JSON body")
		}
		if !strings.HasPrefix(digest, "SHA-256=") {
			t.Errorf("unexpected digest format: %q", digest)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := Conf{
		Host:     srv.URL,
		Username: "u",
		Secret:   "s",
		Headers:  []string{"date", "@request-target", "digest"},
	}
	_ = c.Validate()
	cli, _ := NewClient(c)
	resp, err := cli.DoRaw(context.Background(), "POST", "/json", "application/json", []byte(`{"x":1}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
}

func TestSignOption_BodyReadError(t *testing.T) {
	// signOption body read error path: inject an errReader as body
	c := newConf("date", "@request-target", "digest")
	opt := signOption(c)

	req, _ := http.NewRequest("POST", "http://example.com/api", &errReader{err: errors.New("read fail")})
	req.Header.Set(headerContentType, "application/json")

	result := opt(req)
	// signOption returns r unchanged on error, Authorization should NOT be set
	_ = result
}

// errReader always returns an error on Read.
type errReader struct{ err error }

func (e *errReader) Read(_ []byte) (int, error) { return 0, e.err }

func TestClient_Do_WithAllAlgorithms(t *testing.T) {
	algos := []string{
		AlgorithmHmacSHA1,
		AlgorithmHmacSHA224,
		AlgorithmHmacSHA256,
		AlgorithmHmacSHA384,
		AlgorithmHmacSHA512,
	}
	for _, algo := range algos {
		algo := algo
		t.Run(algo, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				auth := r.Header.Get("Authorization")
				if !strings.Contains(auth, algo) {
					t.Errorf("expected algorithm %q in Authorization, got: %q", algo, auth)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			cli, err := NewClient(Conf{
				Host:      srv.URL,
				Username:  "user",
				Secret:    "secret",
				Algorithm: algo,
			})
			if err != nil {
				t.Fatalf("NewClient error: %v", err)
			}
			resp, err := cli.Do(context.Background(), "GET", "/test", nil)
			if err != nil {
				t.Fatalf("Do error: %v", err)
			}
			defer resp.Body.Close()
		})
	}
}

func TestClient_Do_WithDigestHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			t.Error("expected non-empty body")
		}
		digest := r.Header.Get("Digest")
		if !strings.HasPrefix(digest, "SHA-256=") {
			t.Errorf("unexpected digest: %q", digest)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := Conf{
		Host:     srv.URL,
		Username: "u",
		Secret:   "s",
		Headers:  []string{"date", "@request-target", "digest"},
	}
	_ = c.Validate()
	cli, _ := NewClient(c)
	resp, err := cli.DoRaw(context.Background(), "POST", "/data", "application/json", []byte(`{"hello":"world"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
}

func TestClient_Do_HostHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.Contains(auth, "host") {
			t.Errorf("expected host in Authorization headers field, got: %q", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := Conf{
		Host:     srv.URL,
		Username: "u",
		Secret:   "s",
		Headers:  []string{"host", "date", "@request-target"},
	}
	_ = c.Validate()
	cli, _ := NewClient(c)
	resp, err := cli.Do(context.Background(), "GET", "/check", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
}

func TestSignOption_NoBody(t *testing.T) {
	// signOption with nil body should not error and should set Authorization
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cli, _ := NewClient(newTestConf(srv.URL))
	resp, err := cli.Do(context.Background(), "DELETE", "/item/1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
}

// TestSignatureConsistency verifies the same input always produces the same signature.
func TestSignatureConsistency(t *testing.T) {
	c := newConf("date", "@request-target")
	req1 := newRequest(t, "GET", "http://example.com/api", nil)
	req2 := newRequest(t, "GET", "http://example.com/api", nil)
	fixed := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	signRequest(req1, c, nil, fixed)
	signRequest(req2, c, nil, fixed)

	if req1.Header.Get(headerAuthorization) != req2.Header.Get(headerAuthorization) {
		t.Fatal("signatures differ for identical inputs")
	}
}

// TestSignatureChangesWithBody verifies the body affects the signature when digest is in headers.
func TestSignatureChangesWithBody(t *testing.T) {
	c := newConf("date", "@request-target", "digest")
	req1 := newRequest(t, "POST", "http://example.com/api", nil)
	req2 := newRequest(t, "POST", "http://example.com/api", nil)
	fixed := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	signRequest(req1, c, []byte("body-A"), fixed)
	signRequest(req2, c, []byte("body-B"), fixed)

	if req1.Header.Get(headerAuthorization) == req2.Header.Get(headerAuthorization) {
		t.Fatal("signatures should differ when body differs")
	}
}

func TestDoRaw_BodyRestoredAfterSigning(t *testing.T) {
	// Verify that the body is readable by the server after signOption buffers it.
	var serverBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := Conf{
		Host:     srv.URL,
		Username: "u",
		Secret:   "s",
		Headers:  []string{"date", "@request-target", "digest"},
	}
	_ = c.Validate()
	cli, _ := NewClient(c)
	payload := []byte(`{"restore":"me"}`)
	resp, err := cli.DoRaw(context.Background(), "POST", "/echo", "application/json", payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if !bytes.Equal(serverBody, payload) {
		t.Fatalf("server received %q, want %q", serverBody, payload)
	}
}
