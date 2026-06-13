package gateway

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	headerAccept      = "Accept"
	headerContentMD5  = "Content-MD5"
	headerContentType = "Content-Type"
	headerDate        = "Date"
	headerUserAgent   = "User-Agent"

	headerCaKey              = "X-Ca-Key"
	headerCaNonce            = "X-Ca-Nonce"
	headerCaSignature        = "X-Ca-Signature"
	headerCaSignatureHeaders = "X-Ca-Signature-Headers"
	headerCaSignatureMethod  = "X-Ca-Signature-Method"
	headerCaTimestamp        = "X-Ca-Timestamp"

	defaultAccept     = "*/*"
	defaultUserAgent  = "Go-Aliyun-Gateway-Client"
	defaultSignMethod = "HmacSHA256"
)

// signHeaders are the fixed X-Ca-* headers included in the signature.
// Keys are lowercase per Aliyun API Gateway spec; r.Header.Get still works
// because Go canonicalizes header keys internally during lookup.
var signHeaders = []string{
	"x-ca-key",
	"x-ca-nonce",
	"x-ca-signature-method",
	"x-ca-timestamp",
}

// signRequest populates all required gateway headers and computes the HMAC-SHA256 signature.
// bodyBytes is the raw request body (may be nil), used for Content-MD5 computation.
func signRequest(r *http.Request, appKey, appSecret string, bodyBytes []byte) {
	// Set default headers if not already present.
	if r.Header.Get(headerAccept) == "" {
		r.Header.Set(headerAccept, defaultAccept)
	}
	if r.Header.Get(headerDate) == "" {
		r.Header.Set(headerDate, gmtDate())
	}
	if r.Header.Get(headerUserAgent) == "" {
		r.Header.Set(headerUserAgent, defaultUserAgent)
	}

	// Compute Content-MD5 for non-form bodies.
	// Use HasPrefix to handle Content-Type with charset suffix, e.g.
	// "application/x-www-form-urlencoded; charset=utf-8".
	ct := r.Header.Get(headerContentType)
	if len(bodyBytes) > 0 && !strings.HasPrefix(ct, contentTypeForm) && !strings.HasPrefix(ct, contentTypeMultipart) {
		r.Header.Set(headerContentMD5, md5Hash(bodyBytes))
	}

	// Set mandatory X-Ca-* headers.
	r.Header.Set(headerCaKey, appKey)
	r.Header.Set(headerCaNonce, uuid.New().String())
	r.Header.Set(headerCaSignatureMethod, defaultSignMethod)
	r.Header.Set(headerCaTimestamp, millis())

	// Build string-to-sign and compute signature.
	sts := buildStringToSign(r)
	r.Header.Set(headerCaSignature, hmacSHA256(sts, appSecret))

	// Declare which X-Ca-* headers are signed.
	r.Header.Set(headerCaSignatureHeaders, strings.Join(signHeaders, ","))
}

// buildStringToSign constructs the canonical string per Alibaba Cloud API Gateway v1 spec:
//
//	METHOD\n
//	Accept\n
//	Content-MD5\n
//	Content-Type\n
//	Date\n
//	X-Ca-headers (sorted, key:value)\n
//	URL (path + sorted query)
func buildStringToSign(r *http.Request) string {
	var b strings.Builder

	b.WriteString(r.Method)
	b.WriteByte('\n')
	b.WriteString(r.Header.Get(headerAccept))
	b.WriteByte('\n')
	b.WriteString(r.Header.Get(headerContentMD5))
	b.WriteByte('\n')
	b.WriteString(r.Header.Get(headerContentType))
	b.WriteByte('\n')
	b.WriteString(r.Header.Get(headerDate))
	b.WriteByte('\n')

	// Signed X-Ca-* headers (sorted, lowercase key:value).
	for _, h := range signHeaders {
		b.WriteString(h)
		b.WriteByte(':')
		b.WriteString(r.Header.Get(h))
		b.WriteByte('\n')
	}

	// URL = path + sorted query params.
	b.WriteString(r.URL.Path)
	b.WriteString(sortedQuery(r.URL.RawQuery))

	return b.String()
}

// sortedQuery returns the query string sorted by key for signing.
// Per Aliyun API Gateway spec:
//   - duplicate keys: only the first value is kept
//   - empty values: the '=' is omitted, only the key is included
//
// It operates on the raw (already URL-encoded) query to preserve
// the original encoding and match what the server actually receives.
func sortedQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	seen := make(map[string]bool)
	type kv struct {
		key  string
		sign string
	}
	var pairs []kv

	for _, part := range strings.Split(rawQuery, "&") {
		key, sign := splitQueryPart(part)
		if seen[key] {
			continue
		}
		seen[key] = true
		pairs = append(pairs, kv{key: key, sign: sign})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].key < pairs[j].key
	})

	result := make([]string, len(pairs))
	for i, p := range pairs {
		result[i] = p.sign
	}
	return "?" + strings.Join(result, "&")
}

// splitQueryPart splits a "key=value" or "key" query part.
// Returns the raw key and the signature representation.
// Per Aliyun spec, empty values omit the '=' sign.
func splitQueryPart(part string) (key, sign string) {
	if idx := strings.IndexByte(part, '='); idx >= 0 {
		k := part[:idx]
		v := part[idx+1:]
		if v == "" {
			return k, k
		}
		return k, part
	}
	return part, part
}

// hmacSHA256 computes HMAC-SHA256 and returns base64-encoded result.
func hmacSHA256(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// md5Hash computes MD5 and returns base64-encoded result.
func md5Hash(data []byte) string {
	h := md5.Sum(data)
	return base64.StdEncoding.EncodeToString(h[:])
}

// millis returns the current time in milliseconds as a string.
func millis() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

// gmtDate returns the current time in HTTP GMT date format.
func gmtDate() string {
	return time.Now().UTC().Format(http.TimeFormat)
}
