package minio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client defines the MinIO client interface with convenience methods,
// atomic operations, bucket management, and raw access.
type Client interface {
	// === Convenience methods ===

	// UploadFile uploads a local file to the specified bucket and key.
	UploadFile(ctx context.Context, bucket, key, filePath string, opts ...Option) (*UploadInfo, error)
	// UploadReader uploads data from a reader to the specified bucket and key.
	UploadReader(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts ...Option) (*UploadInfo, error)
	// Download retrieves an object as an io.ReadCloser. Caller must close the reader.
	Download(ctx context.Context, bucket, key string, opts ...Option) (io.ReadCloser, error)
	// Delete removes an object from the specified bucket.
	Delete(ctx context.Context, bucket, key string) error
	// Exists checks if an object exists in the specified bucket.
	Exists(ctx context.Context, bucket, key string) (bool, error)
	// GetPresignedDownloadURL generates a presigned URL for downloading an object.
	GetPresignedDownloadURL(ctx context.Context, bucket, key string, expiry time.Duration, opts ...PresignedOption) (string, error)
	// GetPresignedUploadURL generates a presigned URL for uploading an object.
	GetPresignedUploadURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)

	// === Atomic operations ===

	// PutObject uploads an object with full control over options.
	PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts miniogo.PutObjectOptions) (*UploadInfo, error)
	// GetObject retrieves an object with full control over options.
	GetObject(ctx context.Context, bucket, key string, opts miniogo.GetObjectOptions) (*miniogo.Object, error)
	// StatObject retrieves object metadata.
	StatObject(ctx context.Context, bucket, key string, opts miniogo.StatObjectOptions) (*ObjectInfo, error)
	// RemoveObject deletes an object with full control over options.
	RemoveObject(ctx context.Context, bucket, key string, opts miniogo.RemoveObjectOptions) error
	// CopyObject copies an object from source to destination.
	CopyObject(ctx context.Context, dst miniogo.CopyDestOptions, src miniogo.CopySrcOptions) (*UploadInfo, error)
	// ListObjects returns a channel of objects in the bucket.
	ListObjects(ctx context.Context, bucket string, opts miniogo.ListObjectsOptions) <-chan miniogo.ObjectInfo

	// === Bucket management ===

	// MakeBucket creates a new bucket.
	MakeBucket(ctx context.Context, bucket string, opts miniogo.MakeBucketOptions) error
	// RemoveBucket removes an empty bucket.
	RemoveBucket(ctx context.Context, bucket string) error
	// ListBuckets lists all buckets.
	ListBuckets(ctx context.Context) ([]BucketInfo, error)
	// SetBucketPolicy sets the policy on a bucket.
	SetBucketPolicy(ctx context.Context, bucket, policy string) error
	// GetBucketPolicy gets the policy of a bucket.
	GetBucketPolicy(ctx context.Context, bucket string) (string, error)

	// === Raw access ===

	// RawClient returns the next round-robin selected underlying minio-go client.
	RawClient() *miniogo.Client
	// RawClients returns all underlying minio-go clients.
	RawClients() []*miniogo.Client
}

// Option customizes convenience method behavior.
type Option func(*requestOptions)

// requestOptions holds optional parameters for convenience methods.
type requestOptions struct {
	contentType  string
	metadata     map[string]string
	storageClass string
	partSize     uint64
}

// WithContentType sets the content type for the upload.
func WithContentType(ct string) Option {
	return func(o *requestOptions) {
		o.contentType = ct
	}
}

// WithMetadata sets custom metadata for the upload.
func WithMetadata(m map[string]string) Option {
	return func(o *requestOptions) {
		o.metadata = m
	}
}

// WithStorageClass sets the storage class for the upload.
func WithStorageClass(sc string) Option {
	return func(o *requestOptions) {
		o.storageClass = sc
	}
}

// WithPartSize sets the part size for multipart uploads.
func WithPartSize(size uint64) Option {
	return func(o *requestOptions) {
		o.partSize = size
	}
}

// PresignedOption customizes presigned URL generation.
type PresignedOption func(url.Values)

// WithResponseContentDisposition sets the response Content-Disposition header for presigned URLs.
// Use "inline" for browser preview, "attachment" for forced download.
func WithResponseContentDisposition(disposition string) PresignedOption {
	return func(params url.Values) {
		params.Set("response-content-disposition", disposition)
	}
}

// WithResponseContentType overrides the response Content-Type header for presigned URLs.
func WithResponseContentType(ct string) PresignedOption {
	return func(params url.Values) {
		params.Set("response-content-type", ct)
	}
}

// ClientOption customizes the client during creation.
type ClientOption func(*CommonClient)

// WithTransport sets a custom base http.RoundTripper for all endpoints.
// The instrumented transport will wrap this transport.
func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *CommonClient) {
		c.baseTransport = transport
	}
}

// CommonClient implements Client with P2C load balancing and write-after-read affinity.
type CommonClient struct {
	conf          Conf
	balancer      *p2cBalancer
	baseTransport http.RoundTripper
}

// NewClient creates a MinIO client with P2C load-balanced connections to all endpoints.
func NewClient(c Conf, opts ...ClientOption) (Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	cc := &CommonClient{conf: c}

	// Apply options first to capture baseTransport if provided.
	for _, opt := range opts {
		opt(cc)
	}

	nodes := make([]*node, 0, len(c.Endpoints))
	for _, endpoint := range c.Endpoints {
		transport := newInstrumentedTransport(endpoint, cc.baseTransport, c.SlowThreshold)

		var creds *credentials.Credentials
		if c.SignatureVersion == SignatureV2 {
			creds = credentials.NewStaticV2(c.AccessKeyID, c.SecretAccessKey, "")
		} else {
			creds = credentials.NewStaticV4(c.AccessKeyID, c.SecretAccessKey, "")
		}

		// todo miss some miniogo.Options
		client, err := miniogo.New(endpoint, &miniogo.Options{
			Creds:     creds,
			Secure:    c.UseSSL,
			Region:    c.Region,
			Transport: transport,
		})
		if err != nil {
			return nil, fmt.Errorf("minio: init endpoint %s failed: %w", endpoint, err)
		}
		nodes = append(nodes, &node{client: client, endpoint: endpoint})
	}

	balancer, err := newP2CBalancer(nodes, c.AccessKeyID)
	if err != nil {
		return nil, err
	}
	cc.balancer = balancer

	return cc, nil
}

// MustNewClient creates a Client and panics on error.
func MustNewClient(c Conf, opts ...ClientOption) Client {
	client, err := NewClient(c, opts...)
	logx.Must(err)
	return client
}

// RawClient returns a P2C-selected underlying client.
func (c *CommonClient) RawClient() *miniogo.Client {
	n := c.pickNode()
	if n == nil {
		return nil
	}
	return n.client
}

// RawClients returns all underlying minio-go clients.
func (c *CommonClient) RawClients() []*miniogo.Client {
	clients := make([]*miniogo.Client, len(c.balancer.nodes))
	for i, n := range c.balancer.nodes {
		clients[i] = n.client
	}
	return clients
}

// --- Convenience method implementations ---

// UploadFile uploads a local file using FPutObject with automatic multipart.
func (c *CommonClient) UploadFile(ctx context.Context, bucket, key, filePath string, opts ...Option) (*UploadInfo, error) {
	o := c.buildRequestOptions(opts)
	putOpts := miniogo.PutObjectOptions{
		ContentType:  o.contentType,
		UserMetadata: o.metadata,
		StorageClass: o.storageClass,
		PartSize:     o.partSize,
	}

	return executeWriteWith(c, bucket, key, func(client *miniogo.Client) (*UploadInfo, error) {
		info, err := client.FPutObject(ctx, bucket, key, filePath, putOpts)
		if err != nil {
			return nil, wrapError(err)
		}
		return toUploadInfo(key, info), nil
	})
}

// UploadReader uploads data from a reader.
func (c *CommonClient) UploadReader(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts ...Option) (*UploadInfo, error) {
	o := c.buildRequestOptions(opts)
	putOpts := miniogo.PutObjectOptions{
		ContentType:  o.contentType,
		UserMetadata: o.metadata,
		StorageClass: o.storageClass,
		PartSize:     o.partSize,
	}

	return executeWriteWith(c, bucket, key, func(client *miniogo.Client) (*UploadInfo, error) {
		info, err := client.PutObject(ctx, bucket, key, reader, size, putOpts)
		if err != nil {
			return nil, wrapError(err)
		}
		return toUploadInfo(key, info), nil
	})
}

// Download retrieves an object as an io.ReadCloser.
// Uses affinity-aware selection without failover since it returns a stream.
func (c *CommonClient) Download(ctx context.Context, bucket, key string, opts ...Option) (io.ReadCloser, error) {
	o := c.buildRequestOptions(opts)
	getOpts := miniogo.GetObjectOptions{}
	if o.metadata != nil {
		for k, v := range o.metadata {
			getOpts.Set(k, v)
		}
	}

	n := c.pickNodeWithAffinity(bucket, key)
	if n == nil {
		return nil, fmt.Errorf("minio: no available endpoints")
	}
	obj, err := n.client.GetObject(ctx, bucket, key, getOpts)
	if err != nil {
		return nil, wrapError(err)
	}

	// Verify the object is accessible by reading stat.
	_, err = obj.Stat()
	if err != nil {
		obj.Close()
		return nil, wrapError(err)
	}

	return obj, nil
}

// Delete removes an object from the specified bucket.
func (c *CommonClient) Delete(ctx context.Context, bucket, key string) error {
	return c.execute(func(client *miniogo.Client) error {
		return wrapError(client.RemoveObject(ctx, bucket, key, miniogo.RemoveObjectOptions{}))
	})
}

// Exists checks if an object exists in the specified bucket.
func (c *CommonClient) Exists(ctx context.Context, bucket, key string) (bool, error) {
	return executeWithAffinityWith(c, bucket, key, func(client *miniogo.Client) (bool, error) {
		_, err := client.StatObject(ctx, bucket, key, miniogo.StatObjectOptions{})
		if err != nil {
			var resp miniogo.ErrorResponse
			if miniogo.ToErrorResponse(err) == resp {
				// Check for "NoSuchKey" or 404.
			}
			errResp := miniogo.ToErrorResponse(err)
			if errResp.Code == "NoSuchKey" || errResp.StatusCode == 404 {
				return false, nil
			}
			return false, wrapError(err)
		}
		return true, nil
	})
}

// GetPresignedDownloadURL generates a presigned GET URL.
// Use PresignedOption to control response headers (e.g. WithResponseContentDisposition("inline") for preview).
func (c *CommonClient) GetPresignedDownloadURL(ctx context.Context, bucket, key string, expiry time.Duration, opts ...PresignedOption) (string, error) {
	var reqParams url.Values
	if len(opts) > 0 {
		reqParams = make(url.Values)
		for _, opt := range opts {
			opt(reqParams)
		}
	}
	return executeWith(c, func(client *miniogo.Client) (string, error) {
		u, err := client.PresignedGetObject(ctx, bucket, key, expiry, reqParams)
		if err != nil {
			return "", wrapError(err)
		}
		return u.String(), nil
	})
}

// GetPresignedUploadURL generates a presigned PUT URL.
func (c *CommonClient) GetPresignedUploadURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	return executeWith(c, func(client *miniogo.Client) (string, error) {
		u, err := client.PresignedPutObject(ctx, bucket, key, expiry)
		if err != nil {
			return "", wrapError(err)
		}
		return u.String(), nil
	})
}

// buildRequestOptions applies Option functions and returns the resulting config.
func (c *CommonClient) buildRequestOptions(opts []Option) *requestOptions {
	o := &requestOptions{
		contentType: defaultContentType,
		partSize:    defaultPartSize,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
