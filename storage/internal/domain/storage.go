package domain

import (
	"context"
	"io"
)

// Use context in cloud-based storage to handle timeouts and cancellations.
type Storage interface {
	Save(ctx context.Context, path string) (stream io.WriteCloser, err error)
	Get(ctx context.Context, path string) (stream io.ReadCloser, err error)
}
