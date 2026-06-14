package oss

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/rest/httpc"
)

// signOption returns an httpc.Option that signs every outgoing request
// using the Alibaba Cloud OSS V1 HMAC-SHA1 algorithm.
func signOption(c Conf) httpc.Option {
	secret := []byte(c.AccessKeySecret)
	bucket := c.Bucket
	return func(r *http.Request) *http.Request {
		// Set default headers if not already present.
		if r.Header.Get(headerDate) == "" {
			r.Header.Set(headerDate, time.Now().UTC().Format(http.TimeFormat))
		}
		if r.Header.Get(headerUserAgent) == "" {
			r.Header.Set(headerUserAgent, defaultUserAgent)
		}

		// Build CanonicalizedResource: /{bucket}/{path}?{sub-resources}
		resource := "/" + bucket
		if r.URL.Path != "/" && r.URL.Path != "" {
			resource += r.URL.Path
		}
		if sub := canonicalizedSubResources(r.URL.Query()); sub != "" {
			resource += "?" + sub
		}

		sts := buildStringToSign(r, resource)
		sig := hmacSHA1(sts, secret)
		r.Header.Set(headerAuthorization, authPrefix+c.AccessKeyId+":"+sig)

		return r
	}
}

// buildStringToSign constructs the canonical string per Alibaba Cloud OSS V1 spec:
//
//	HTTP-Verb + "\n" +
//	Content-MD5 + "\n" +
//	Content-Type + "\n" +
//	Date + "\n" +
//	CanonicalizedLOSSHeaders +
//	CanonicalizedResource
func buildStringToSign(r *http.Request, canonicalizedResource string) string {
	var b strings.Builder

	b.WriteString(r.Method)
	b.WriteByte('\n')
	b.WriteString(r.Header.Get(headerContentMD5))
	b.WriteByte('\n')
	b.WriteString(r.Header.Get(headerContentType))
	b.WriteByte('\n')
	b.WriteString(r.Header.Get(headerDate))
	b.WriteByte('\n')

	// CanonicalizedLOSSHeaders: x-oss-* headers sorted by lowercase key.
	b.WriteString(canonicalizedOSSHeaders(r.Header))

	// CanonicalizedResource
	b.WriteString(canonicalizedResource)

	return b.String()
}

// canonicalizedLOSSHeaders collects all x-oss-* request headers,
// sorts them by lowercase key, and formats them as "key:value\n".
func canonicalizedOSSHeaders(headers http.Header) string {
	var keys []string
	for k := range headers {
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, "x-oss-") {
			keys = append(keys, lk)
		}
	}
	if len(keys) == 0 {
		return ""
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte(':')
		b.WriteString(headers.Get(k))
		b.WriteByte('\n')
	}
	return b.String()
}

// canonicalizedSubResources extracts OSS-defined sub-resources from query params,
// sorts them by key, and returns the canonical "?k1=v1&k2=v2" string (without leading ?).
// Empty values are represented as just the key (no "=").
// Regular query parameters (prefix, marker, max-keys, etc.) are excluded.
func canonicalizedSubResources(query url.Values) string {
	var keys []string
	for k := range query {
		if _, ok := ossSubResources[k]; ok {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return ""
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := query.Get(k)
		if v == "" {
			parts = append(parts, k)
		} else {
			parts = append(parts, k+"="+v)
		}
	}
	return strings.Join(parts, "&")
}

// hmacSHA1 computes HMAC-SHA1 and returns base64-encoded result.
func hmacSHA1(data string, key []byte) string {
	h := hmac.New(sha1.New, key)
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
