package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage provides MinIO object storage operations
type Storage struct {
	client     *minio.Client
	bucket     string
	urlExpiry  time.Duration
}

// Config holds MinIO configuration
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	URLExpiry time.Duration
}

// DefaultConfig returns default MinIO configuration
func DefaultConfig() Config {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadmin123"
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "sovereign-firm"
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	return Config{
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Bucket:    bucket,
		UseSSL:    useSSL,
		URLExpiry: 1 * time.Hour,
	}
}

// New creates a new MinIO storage client
func New(ctx context.Context, cfg Config) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	// Ensure bucket exists
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}

	return &Storage{
		client:    client,
		bucket:    cfg.Bucket,
		urlExpiry: cfg.URLExpiry,
	}, nil
}

// Client returns the underlying MinIO client
func (s *Storage) Client() *minio.Client {
	return s.client
}

// Bucket returns the bucket name
func (s *Storage) Bucket() string {
	return s.bucket
}

// PutObject uploads an object
func (s *Storage) PutObject(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error {
	opts := minio.PutObjectOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}

	_, err := s.client.PutObject(ctx, s.bucket, path, reader, size, opts)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	return nil
}

// GetObject retrieves an object
func (s *Storage) GetObject(ctx context.Context, path string) (io.ReadCloser, *minio.ObjectInfo, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("get object: %w", err)
	}

	info, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, nil, fmt.Errorf("stat object: %w", err)
	}

	return obj, &info, nil
}

// DeleteObject removes an object
func (s *Storage) DeleteObject(ctx context.Context, path string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	return nil
}

// ListObjects lists objects with a prefix
func (s *Storage) ListObjects(ctx context.Context, prefix string) ([]minio.ObjectInfo, error) {
	var objects []minio.ObjectInfo

	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for obj := range objectCh {
		if obj.Err != nil {
			return nil, fmt.Errorf("list objects: %w", obj.Err)
		}
		objects = append(objects, obj)
	}

	return objects, nil
}

// GetPresignedURL generates a presigned URL for downloading
func (s *Storage) GetPresignedURL(ctx context.Context, path string) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, path, s.urlExpiry, nil)
	if err != nil {
		return "", fmt.Errorf("presign url: %w", err)
	}
	return url.String(), nil
}

// GetPresignedUploadURL generates a presigned URL for uploading
func (s *Storage) GetPresignedUploadURL(ctx context.Context, path string) (string, error) {
	url, err := s.client.PresignedPutObject(ctx, s.bucket, path, s.urlExpiry)
	if err != nil {
		return "", fmt.Errorf("presign upload url: %w", err)
	}
	return url.String(), nil
}

// ObjectExists checks if an object exists
func (s *Storage) ObjectExists(ctx context.Context, path string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CopyObject copies an object within the bucket
func (s *Storage) CopyObject(ctx context.Context, srcPath, dstPath string) error {
	src := minio.CopySrcOptions{
		Bucket: s.bucket,
		Object: srcPath,
	}
	dst := minio.CopyDestOptions{
		Bucket: s.bucket,
		Object: dstPath,
	}
	_, err := s.client.CopyObject(ctx, dst, src)
	if err != nil {
		return fmt.Errorf("copy object: %w", err)
	}
	return nil
}

// Health checks MinIO health
func (s *Storage) Health(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucket)
	return err
}
