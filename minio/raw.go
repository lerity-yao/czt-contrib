package minio

import (
	"context"
	"fmt"
	"io"

	miniogo "github.com/minio/minio-go/v7"
)

// --- Atomic operations ---

// PutObject uploads an object with full control over PutObjectOptions.
func (c *CommonClient) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts miniogo.PutObjectOptions) (*UploadInfo, error) {
	return executeWriteWith(c, bucket, key, func(client *miniogo.Client) (*UploadInfo, error) {
		info, err := client.PutObject(ctx, bucket, key, reader, size, opts)
		if err != nil {
			return nil, wrapError(err)
		}
		return toUploadInfo(key, info), nil
	})
}

// GetObject retrieves an object. Uses affinity-aware selection without failover
// since it returns a streaming object.
func (c *CommonClient) GetObject(ctx context.Context, bucket, key string, opts miniogo.GetObjectOptions) (*miniogo.Object, error) {
	n := c.pickNodeWithAffinity(bucket, key)
	if n == nil {
		return nil, fmt.Errorf("minio: no available endpoints")
	}
	obj, err := n.client.GetObject(ctx, bucket, key, opts)
	if err != nil {
		return nil, wrapError(err)
	}
	return obj, nil
}

// StatObject retrieves object metadata.
func (c *CommonClient) StatObject(ctx context.Context, bucket, key string, opts miniogo.StatObjectOptions) (*ObjectInfo, error) {
	return executeWithAffinityWith(c, bucket, key, func(client *miniogo.Client) (*ObjectInfo, error) {
		info, err := client.StatObject(ctx, bucket, key, opts)
		if err != nil {
			return nil, wrapError(err)
		}
		return toObjectInfo(info), nil
	})
}

// RemoveObject deletes an object with full control over RemoveObjectOptions.
func (c *CommonClient) RemoveObject(ctx context.Context, bucket, key string, opts miniogo.RemoveObjectOptions) error {
	return c.execute(func(client *miniogo.Client) error {
		return wrapError(client.RemoveObject(ctx, bucket, key, opts))
	})
}

// CopyObject copies an object from source to destination.
func (c *CommonClient) CopyObject(ctx context.Context, dst miniogo.CopyDestOptions, src miniogo.CopySrcOptions) (*UploadInfo, error) {
	return executeWriteWith(c, dst.Bucket, dst.Object, func(client *miniogo.Client) (*UploadInfo, error) {
		info, err := client.CopyObject(ctx, dst, src)
		if err != nil {
			return nil, wrapError(err)
		}
		return toUploadInfo(dst.Object, info), nil
	})
}

// ListObjects returns a channel of objects in the bucket.
// Uses P2C selection without failover since it returns a channel.
func (c *CommonClient) ListObjects(ctx context.Context, bucket string, opts miniogo.ListObjectsOptions) <-chan miniogo.ObjectInfo {
	n := c.pickNode()
	if n == nil {
		ch := make(chan miniogo.ObjectInfo)
		close(ch)
		return ch
	}
	return n.client.ListObjects(ctx, bucket, opts)
}

// --- Bucket management ---

// MakeBucket creates a new bucket.
func (c *CommonClient) MakeBucket(ctx context.Context, bucket string, opts miniogo.MakeBucketOptions) error {
	return c.execute(func(client *miniogo.Client) error {
		return wrapError(client.MakeBucket(ctx, bucket, opts))
	})
}

// RemoveBucket removes an empty bucket.
func (c *CommonClient) RemoveBucket(ctx context.Context, bucket string) error {
	return c.execute(func(client *miniogo.Client) error {
		return wrapError(client.RemoveBucket(ctx, bucket))
	})
}

// ListBuckets lists all buckets.
func (c *CommonClient) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	return executeWith(c, func(client *miniogo.Client) ([]BucketInfo, error) {
		buckets, err := client.ListBuckets(ctx)
		if err != nil {
			return nil, wrapError(err)
		}
		result := make([]BucketInfo, 0, len(buckets))
		for _, b := range buckets {
			result = append(result, BucketInfo{
				Name:         b.Name,
				CreationDate: b.CreationDate,
			})
		}
		return result, nil
	})
}

// SetBucketPolicy sets the access policy on a bucket.
func (c *CommonClient) SetBucketPolicy(ctx context.Context, bucket, policy string) error {
	return c.execute(func(client *miniogo.Client) error {
		return wrapError(client.SetBucketPolicy(ctx, bucket, policy))
	})
}

// GetBucketPolicy gets the access policy of a bucket.
func (c *CommonClient) GetBucketPolicy(ctx context.Context, bucket string) (string, error) {
	return executeWith(c, func(client *miniogo.Client) (string, error) {
		policy, err := client.GetBucketPolicy(ctx, bucket)
		if err != nil {
			return "", wrapError(err)
		}
		return policy, nil
	})
}
