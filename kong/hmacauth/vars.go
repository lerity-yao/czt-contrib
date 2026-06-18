package hmacauth

import "strings"

// Standard HTTP headers used in signing.
const (
	headerAuthorization = "Authorization"
	headerContentType   = "Content-Type"
	headerDate          = "Date"
	headerDigest        = "Digest"
	headerHost          = "Host"
	headerUserAgent     = "User-Agent"
)

// pseudoRequestTarget is the pseudo-header recognized by Kong HMAC Auth
// for the request target line. Value format: "method /path?query" (lowercase method).
const pseudoRequestTarget = "@request-target"

// Supported HMAC algorithms.
const (
	AlgorithmHmacSHA1   = "hmac-sha1"   // SHA-1, disabled by default in Kong
	AlgorithmHmacSHA224 = "hmac-sha224" // requires Kong 3.14+
	AlgorithmHmacSHA256 = "hmac-sha256" // recommended default
	AlgorithmHmacSHA384 = "hmac-sha384"
	AlgorithmHmacSHA512 = "hmac-sha512"
)

// defaultAlgorithm is used when Conf.Algorithm is empty.
const defaultAlgorithm = AlgorithmHmacSHA256

// defaultHeaders is the default set of headers included in the signing string
// when Conf.Headers is not specified.
var defaultHeaders = []string{
	"date",
	pseudoRequestTarget,
}

// defaultSignHeaders returns a copy of the default headers slice.
func defaultSignHeaders() []string {
	cp := make([]string, len(defaultHeaders))
	copy(cp, defaultHeaders)
	return cp
}

// defaultUserAgent is sent on every request.
const defaultUserAgent = "Go-Kong-HmacAuth-Client"

// digestPrefix is the prefix for the Digest header value per RFC 3230.
const digestPrefix = "SHA-256="

// contentTypeForm and contentTypeMultipart are used to skip body digest for form payloads.
const (
	contentTypeForm      = "application/x-www-form-urlencoded"
	contentTypeMultipart = "multipart/form-data"
)

// validAlgorithms is the set of algorithms accepted by this SDK.
var validAlgorithms = map[string]bool{
	AlgorithmHmacSHA1:   true,
	AlgorithmHmacSHA224: true,
	AlgorithmHmacSHA256: true,
	AlgorithmHmacSHA384: true,
	AlgorithmHmacSHA512: true,
}

// joinHeaders returns headers joined by a single space, as required by the
// Authorization header's headers field.
func joinHeaders(headers []string) string {
	return strings.Join(headers, " ")
}
