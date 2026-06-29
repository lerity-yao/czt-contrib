package minio

import (
	"errors"
	"testing"
	"time"

	miniogo "github.com/minio/minio-go/v7"
)

// --- Error type tests ---

func TestError_WithKeyAndBucket(t *testing.T) {
	e := &Error{Code: "NoSuchKey", Message: "not found", BucketName: "bkt", Key: "k", StatusCode: 404}
	s := e.Error()
	if s == "" {
		t.Fatal("expected non-empty error string")
	}
	if !errors.Is(e, e) {
		t.Fatal("expected errors.Is to match self")
	}
}

func TestError_BucketOnly(t *testing.T) {
	e := &Error{Code: "NoSuchBucket", Message: "not found", BucketName: "bkt", StatusCode: 404}
	s := e.Error()
	if s == "" {
		t.Fatal("expected non-empty error string")
	}
}

func TestError_NoBucketNoKey(t *testing.T) {
	e := &Error{Code: "InternalError", Message: "oops", StatusCode: 500}
	s := e.Error()
	if s == "" {
		t.Fatal("expected non-empty error string")
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	e := &Error{Code: "X", Message: "m", Cause: cause}
	if e.Unwrap() != cause {
		t.Fatal("Unwrap should return cause")
	}
}

func TestError_Unwrap_Nil(t *testing.T) {
	e := &Error{Code: "X", Message: "m"}
	if e.Unwrap() != nil {
		t.Fatal("Unwrap should return nil when no cause")
	}
}

// --- wrapError tests ---

func TestWrapError_Nil(t *testing.T) {
	if wrapError(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestWrapError_MinioErrorResponse(t *testing.T) {
	resp := miniogo.ErrorResponse{
		Code:       "NoSuchKey",
		Message:    "not found",
		BucketName: "bkt",
		Key:        "key1",
		StatusCode: 404,
	}
	err := wrapError(resp)
	var e *Error
	if !errors.As(err, &e) {
		t.Fatal("expected *Error type")
	}
	if e.Code != "NoSuchKey" {
		t.Fatalf("expected 'NoSuchKey', got '%s'", e.Code)
	}
	if e.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", e.StatusCode)
	}
}

func TestWrapError_NonMinioError(t *testing.T) {
	orig := errors.New("some error")
	err := wrapError(orig)
	if err != orig {
		t.Fatal("expected original error returned as-is")
	}
}

// --- toUploadInfo tests ---

func TestToUploadInfo(t *testing.T) {
	info := toUploadInfo("mykey", miniogo.UploadInfo{
		ETag:      "etag1",
		Size:      100,
		VersionID: "v1",
	})
	if info.Key != "mykey" {
		t.Fatalf("expected key='mykey', got '%s'", info.Key)
	}
	if info.ETag != "etag1" {
		t.Fatalf("expected etag='etag1', got '%s'", info.ETag)
	}
	if info.Size != 100 {
		t.Fatalf("expected size=100, got %d", info.Size)
	}
	if info.VersionID != "v1" {
		t.Fatalf("expected version='v1', got '%s'", info.VersionID)
	}
}

// --- toObjectInfo tests ---

func TestToObjectInfo(t *testing.T) {
	now := time.Now()
	info := toObjectInfo(miniogo.ObjectInfo{
		Key:          "obj1",
		Size:         256,
		ContentType:  "text/plain",
		ETag:         "etag2",
		LastModified: now,
		UserMetadata: map[string]string{"author": "test"},
	})
	if info.Key != "obj1" {
		t.Fatalf("expected key='obj1', got '%s'", info.Key)
	}
	if info.Size != 256 {
		t.Fatalf("expected size=256, got %d", info.Size)
	}
	if info.ContentType != "text/plain" {
		t.Fatalf("expected 'text/plain', got '%s'", info.ContentType)
	}
	if info.Metadata["author"] != "test" {
		t.Fatal("expected metadata with author=test")
	}
}

func TestToObjectInfo_NoMetadata(t *testing.T) {
	info := toObjectInfo(miniogo.ObjectInfo{
		Key:  "obj2",
		Size: 0,
	})
	if info.Key != "obj2" {
		t.Fatal("expected key='obj2'")
	}
	if info.Metadata == nil {
		t.Fatal("metadata should be non-nil empty map")
	}
	if len(info.Metadata) != 0 {
		t.Fatal("expected empty metadata")
	}
}
