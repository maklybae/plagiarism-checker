package wordcloud

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

func (s *Storage) Save(_ context.Context, path string) (io.WriteCloser, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("wordcloud storage: failed to create file %q: %w", path, err)
	}

	return file, nil
}

func (s *Storage) Get(_ context.Context, path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("wordcloud storage: failed to open file %q: %w", path, err)
	}

	return file, nil
}
