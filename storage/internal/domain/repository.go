package domain

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateFile(ctx context.Context, file *FileInfo) error
	GetFileByID(ctx context.Context, id uuid.UUID) (file *FileInfo, err error)
	GetFilesByHash(ctx context.Context, hash string) ([]*FileInfo, error)
}
