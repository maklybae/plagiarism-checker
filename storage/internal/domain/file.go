package domain

import "github.com/google/uuid"

type FileData []byte

type FileInfo struct {
	ID   uuid.UUID
	Path string
	Hash string
}
