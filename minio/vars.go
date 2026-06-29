package minio

// Default content type for uploaded objects.
const defaultContentType = "application/octet-stream"

// Default part size for multipart uploads (16 MiB).
const defaultPartSize = 16 * 1024 * 1024 // 16MB

// Minimum part size allowed by MinIO (5 MiB).
const minPartSize = 5 * 1024 * 1024 // 5MB

// Signature version constants.
const (
	// SignatureV2 uses the legacy V2 signing algorithm.
	SignatureV2 = "v2"
	// SignatureV4 uses the recommended V4 signing algorithm.
	SignatureV4 = "v4"
)
