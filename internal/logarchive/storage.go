package logarchive

import "context"

// StorageClient abstracts object storage for log chunks.
// Implementations may use MinIO or a mock for tests.
type StorageClient interface {
	PutObject(ctx context.Context, bucket, key string, data []byte, contentType string) error
	GetObject(ctx context.Context, bucket, key string) ([]byte, error)
	ListObjects(ctx context.Context, bucket, prefix string) ([]string, error)
}
