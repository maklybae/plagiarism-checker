package ctxio

import (
	"context"
	"io"
)

type ContextReader struct {
	r   io.Reader
	ctx context.Context
}

func NewContextReader(ctx context.Context, r io.Reader) *ContextReader {
	return &ContextReader{
		r:   r,
		ctx: ctx,
	}
}

func (cr *ContextReader) Read(p []byte) (n int, err error) {
	select {
	case <-cr.ctx.Done():
		return 0, cr.ctx.Err()
	default:
		return cr.r.Read(p)
	}
}
