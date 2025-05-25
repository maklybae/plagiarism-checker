package local

import (
	"context"
	"fmt"
	"io"
	"os"
)

type Storage struct{}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Save(_ context.Context, path string) (stream io.WriteCloser, err error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", path, err)
	}

	return file, nil
}

func (s *Storage) Get(_ context.Context, path string) (stream io.ReadCloser, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}

	return file, nil
}
