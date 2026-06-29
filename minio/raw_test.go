package minio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	miniogo "github.com/minio/minio-go/v7"
)

// --- Atomic operations tests ---

func TestPutObject_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			io.Copy(io.Discard, r.Body)
			w.Header().Set("ETag", `"put-etag"`)
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	reader := strings.NewReader("test data")
	info, err := c.PutObject(context.Background(), "bucket", "key", reader, 9, miniogo.PutObjectOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil upload info")
	}
	if info.Key != "key" {
		t.Fatalf("expected key='key', got '%s'", info.Key)
	}
}

func TestPutObject_Error(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><Error><Code>AccessDenied</Code><Message>Denied</Message></Error>`)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	reader := strings.NewReader("test data")
	_, err := c.PutObject(context.Background(), "bucket", "key", reader, 9, miniogo.PutObjectOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetObject_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "4")
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("ETag", `"get-etag"`)
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Last-Modified", "Mon, 29 Jun 2026 00:00:00 GMT")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data"))
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	obj, err := c.GetObject(context.Background(), "bucket", "key", miniogo.GetObjectOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj == nil {
		t.Fatal("expected non-nil object")
	}
	obj.Close()
}

func TestGetObject_NoEndpoints(t *testing.T) {
	cc := newTestCommonClient(nil)
	_, err := cc.GetObject(context.Background(), "bucket", "key", miniogo.GetObjectOptions{})
	if err == nil {
		t.Fatal("expected error for no endpoints")
	}
}

func TestStatObject_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "42")
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("ETag", `"stat-etag"`)
			w.Header().Set("Last-Modified", "Mon, 29 Jun 2026 00:00:00 GMT")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	info, err := c.StatObject(context.Background(), "bucket", "key", miniogo.StatObjectOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil object info")
	}
	if info.Size != 42 {
		t.Fatalf("expected size=42, got %d", info.Size)
	}
}

func TestStatObject_NotFound(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	_, err := c.StatObject(context.Background(), "bucket", "nokey", miniogo.StatObjectOptions{})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestRemoveObject_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	err := c.RemoveObject(context.Background(), "bucket", "key", miniogo.RemoveObjectOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCopyObject_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.Header.Get("X-Amz-Copy-Source") != "" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><CopyObjectResult><ETag>"copy-etag"</ETag><LastModified>2026-06-29T00:00:00Z</LastModified></CopyObjectResult>`)
			return
		}
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "10")
			w.Header().Set("ETag", `"src-etag"`)
			w.Header().Set("Last-Modified", "Mon, 29 Jun 2026 00:00:00 GMT")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	dst := miniogo.CopyDestOptions{Bucket: "dstbucket", Object: "dstkey"}
	src := miniogo.CopySrcOptions{Bucket: "srcbucket", Object: "srckey"}
	info, err := c.CopyObject(context.Background(), dst, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil upload info")
	}
	if info.Key != "dstkey" {
		t.Fatalf("expected key='dstkey', got '%s'", info.Key)
	}
}

func TestListObjects_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><ListBucketResult><Name>bucket</Name><IsTruncated>false</IsTruncated><Contents><Key>obj1</Key><Size>10</Size></Contents></ListBucketResult>`)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	ch := c.ListObjects(context.Background(), "bucket", miniogo.ListObjectsOptions{})
	var objects []miniogo.ObjectInfo
	for obj := range ch {
		objects = append(objects, obj)
	}
	if len(objects) == 0 {
		t.Fatal("expected at least one object")
	}
}

func TestListObjects_NoEndpoints(t *testing.T) {
	cc := newTestCommonClient(nil)
	ch := cc.ListObjects(context.Background(), "bucket", miniogo.ListObjectsOptions{})
	count := 0
	for range ch {
		count++
	}
	if count != 0 {
		t.Fatal("expected empty channel for no endpoints")
	}
}

// --- Bucket management tests ---

func TestMakeBucket_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	err := c.MakeBucket(context.Background(), "newbucket", miniogo.MakeBucketOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveBucket_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	err := c.RemoveBucket(context.Background(), "oldbucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListBuckets_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><ListAllMyBucketsResult><Buckets><Bucket><Name>bucket1</Name><CreationDate>2026-01-01T00:00:00Z</CreationDate></Bucket><Bucket><Name>bucket2</Name><CreationDate>2026-06-01T00:00:00Z</CreationDate></Bucket></Buckets></ListAllMyBucketsResult>`)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	buckets, err := c.ListBuckets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}
	if buckets[0].Name != "bucket1" {
		t.Fatalf("expected 'bucket1', got '%s'", buckets[0].Name)
	}
}

func TestSetBucketPolicy_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && strings.Contains(r.URL.RawQuery, "policy") {
			io.Copy(io.Discard, r.Body)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	err := c.SetBucketPolicy(context.Background(), "bucket", `{"Version":"2012-10-17"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetBucketPolicy_Success(t *testing.T) {
	server := newFakeMinioServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.Contains(r.URL.RawQuery, "policy") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"Version":"2012-10-17","Statement":[]}`)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c, _ := NewClient(validConf(hostFromServer(server)))
	policy, err := c.GetBucketPolicy(context.Background(), "bucket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy == "" {
		t.Fatal("expected non-empty policy")
	}
}
