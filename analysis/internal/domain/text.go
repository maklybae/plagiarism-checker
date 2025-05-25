package domain

import "github.com/google/uuid"

type TextInfo struct {
	ID            uuid.UUID
	Lines         int
	Words         int
	Chars         int
	WordcloudPath string
}
