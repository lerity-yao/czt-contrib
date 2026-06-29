package minio

import (
	"errors"
	"fmt"
	"time"

	miniogo "github.com/minio/minio-go/v7"
)

// UploadInfo holds the result of an upload operation.
type UploadInfo struct {
	Key       string
	ETag      string
	Size      int64
	VersionID string
}

// ObjectInfo holds metadata about an object.
type ObjectInfo struct {
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
	Metadata     map[string]string
}

// BucketInfo holds information about a bucket.
type BucketInfo struct {
	Name         string
	CreationDate time.Time
}

// Error is a custom error type that wraps MinIO error details.
type Error struct {
	Code       string // MinIO error code
	Message    string
	BucketName string
	Key        string
	StatusCode int
	Cause      error
}

// Error returns a formatted error string.
func (e *Error) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("minio: %s — %s (bucket: %s, key: %s, status: %d)",
			e.Code, e.Message, e.BucketName, e.Key, e.StatusCode)
	}
	if e.BucketName != "" {
		return fmt.Sprintf("minio: %s — %s (bucket: %s, status: %d)",
			e.Code, e.Message, e.BucketName, e.StatusCode)
	}
	return fmt.Sprintf("minio: %s — %s (status: %d)", e.Code, e.Message, e.StatusCode)
}

// Unwrap returns the underlying cause error.
func (e *Error) Unwrap() error { return e.Cause }

// wrapError converts a minio-go ErrorResponse into a custom Error.
// If the error is not an ErrorResponse, it is returned as-is.
func wrapError(err error) error {
	if err == nil {
		return nil
	}

	var resp miniogo.ErrorResponse
	if errors.As(err, &resp) {
		return &Error{
			Code:       resp.Code,
			Message:    resp.Message,
			BucketName: resp.BucketName,
			Key:        resp.Key,
			StatusCode: resp.StatusCode,
			Cause:      err,
		}
	}

	return err
}

// toUploadInfo converts minio-go UploadInfo to our UploadInfo type.
func toUploadInfo(key string, info miniogo.UploadInfo) *UploadInfo {
	return &UploadInfo{
		Key:       key,
		ETag:      info.ETag,
		Size:      info.Size,
		VersionID: info.VersionID,
	}
}

// toObjectInfo converts minio-go ObjectInfo to our ObjectInfo type.
func toObjectInfo(info miniogo.ObjectInfo) *ObjectInfo {
	metadata := make(map[string]string, len(info.UserMetadata))
	for k, v := range info.UserMetadata {
		metadata[k] = v
	}
	return &ObjectInfo{
		Key:          info.Key,
		Size:         info.Size,
		ContentType:  info.ContentType,
		ETag:         info.ETag,
		LastModified: info.LastModified,
		Metadata:     metadata,
	}
}
