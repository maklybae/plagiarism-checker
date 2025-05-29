package application

import (
	"fmt"

	"github.com/google/uuid"
)

type FileNotFoundError struct {
	ID  uuid.UUID
	Err error
}

func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("file with ID %s not found: %v", e.ID, e.Err)
}

func (e *FileNotFoundError) Unwrap() error {
	return e.Err
}
