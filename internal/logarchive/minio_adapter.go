package logarchive

import (
	"context"
	"bytes"
	"io"

	"github.com/minio/minio-go/v7"
)

const logBucket = "zcid-logs"

// MinIOAdapter implements StorageClient using MinIO.
type MinIOAdapter struct {
	client *minio.Client
}

// NewMinIOAdapter creates an adapter for the MinIO client.
func NewMinIOAdapter(client *minio.Client) *MinIOAdapter {
	return &MinIOAdapter{client: client}
}

// PutObject uploads data to the given bucket and key.
func (a *MinIOAdapter) PutObject(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	_, err := a.client.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	return err
}

// GetObject downloads object content.
func (a *MinIOAdapter) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	obj, err := a.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

// ListObjects returns object keys under the prefix.
func (a *MinIOAdapter) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	ch := a.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true})
	var keys []string
	for obj := range ch {
		if obj.Err != nil {
			return nil, obj.Err
		}
		keys = append(keys, obj.Key)
	}
	return keys, nil
}
