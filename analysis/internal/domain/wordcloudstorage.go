package domain

import (
	"context"
	"io"
)

type WordcloudStorage interface {
	Save(ctx context.Context, path string) (stream io.WriteCloser, err error)
	Get(ctx context.Context, path string) (stream io.ReadCloser, err error)
}
