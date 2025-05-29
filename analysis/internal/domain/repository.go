package domain

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateTextInfo(ctx context.Context, text *TextInfo) error
	GetTextInfoByID(ctx context.Context, id uuid.UUID) (*TextInfo, error)
	SetWordcloudPath(ctx context.Context, id uuid.UUID, path string) error
}
