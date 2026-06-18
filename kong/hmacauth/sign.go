package hmacauth

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"net/http"
	"strings"
	"time"
)

// signRequest injects all Kong HMAC Auth headers and sets the Authorization header.
// bodyBytes is the raw request body; may be nil for requests without a body.
// now is the time basis used for the Date header.
func signRequest(r *http.Request, c Conf, bodyBytes []byte, now time.Time) {
	// Inject Date if it will be part of the signing string.
	if includesHeader(c.Headers, "date") && r.Header.Get(headerDate) == "" {
		r.Header.Set(headerDate, now.UTC().Format(http.TimeFormat))
	}

	// Inject User-Agent if not already set.
	if r.Header.Get(headerUserAgent) == "" {
		r.Header.Set(headerUserAgent, defaultUserAgent)
	}

	// Inject Host if it will be part of the signing string.
	// Go stores Host in r.Host, not in the Header map. We set it explicitly
	// to ensure the signed value matches the value actually sent by http.Client.
	if includesHeader(c.Headers, "host") && r.Header.Get(headerHost) == "" {
		r.Header.Set(headerHost, r.Host)
	}

	// Inject Digest if digest is in the signing headers.
	// Per Kong spec, empty body should also have a Digest (SHA-256 of zero-length body).
	// sha256.Sum256(nil) correctly produces the zero-length body digest.
	if includesHeader(c.Headers, "digest") {
		ct := r.Header.Get(headerContentType)
		isForm := strings.HasPrefix(ct, contentTypeForm) || strings.HasPrefix(ct, contentTypeMultipart)
		if !isForm {
			h := sha256.Sum256(bodyBytes)
			r.Header.Set(headerDigest, digestPrefix+base64.StdEncoding.EncodeToString(h[:]))
		}
	}

	// Build signing string.
	sts := buildSigningString(r, c)

	// Compute HMAC signature.
	sig := computeHMAC(c.Algorithm, []byte(c.Secret), sts)

	// Build Authorization header value.
	r.Header.Set(headerAuthorization, buildAuthorization(c, sig))
}

// buildSigningString constructs the canonical signing string per Kong HMAC Auth spec.
// Each header line is: "header-name: value\n" (note: lowercase header name, original value).
// The last line has no trailing newline.
func buildSigningString(r *http.Request, c Conf) string {
	var b strings.Builder
	for i, h := range c.Headers {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(h)
		b.WriteString(": ")
		b.WriteString(headerValue(r, h))
	}
	return b.String()
}

// headerValue resolves the value for a given header name (or pseudo-header).
func headerValue(r *http.Request, h string) string {
	switch h {
	case pseudoRequestTarget:
		// "@request-target: method /path?query" — all lowercase method.
		return strings.ToLower(r.Method) + " " + r.URL.RequestURI()
	case "host":
		// Go stores Host in r.Host, not in the Header map for client requests.
		// Check the Header map first (in case it was explicitly set), then fall back to r.Host.
		if v := r.Header.Get(headerHost); v != "" {
			return v
		}
		return r.Host
	default:
		return r.Header.Get(http.CanonicalHeaderKey(h))
	}
}

// includesHeader checks whether name (lowercase) appears in headers.
func includesHeader(headers []string, name string) bool {
	for _, h := range headers {
		if h == name {
			return true
		}
	}
	return false
}

// computeHMAC computes the HMAC for the given algorithm and returns base64-encoded result.
func computeHMAC(algorithm string, key []byte, data string) string {
	var h hash.Hash
	switch algorithm {
	case AlgorithmHmacSHA1:
		h = hmac.New(sha1.New, key)
	case AlgorithmHmacSHA224:
		h = hmac.New(sha256.New224, key)
	case AlgorithmHmacSHA384:
		h = hmac.New(sha512.New384, key)
	case AlgorithmHmacSHA512:
		h = hmac.New(sha512.New, key)
	default: // hmac-sha256
		h = hmac.New(sha256.New, key)
	}
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// buildAuthorization assembles the Kong HMAC Authorization header value.
// Format: hmac username="...", algorithm="...", headers="...", signature="..."
func buildAuthorization(c Conf, signature string) string {
	var b strings.Builder
	b.WriteString("hmac ")
	b.WriteString(fmt.Sprintf(`username="%s"`, c.Username))
	b.WriteString(fmt.Sprintf(`, algorithm="%s"`, c.Algorithm))
	b.WriteString(fmt.Sprintf(`, headers="%s"`, joinHeaders(c.Headers)))
	b.WriteString(fmt.Sprintf(`, signature="%s"`, signature))
	return b.String()
}
