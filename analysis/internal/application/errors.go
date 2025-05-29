package application

import (
	"fmt"

	"github.com/google/uuid"
)

type FileInfoNotFoundError struct {
	ID  uuid.UUID
	Err error
}

func (e *FileInfoNotFoundError) Error() string {
	return fmt.Sprintf("file info with ID %s not found: %v", e.ID, e.Err)
}

func (e *FileInfoNotFoundError) Unwrap() error {
	return e.Err
}
