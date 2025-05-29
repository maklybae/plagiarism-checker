package application

import (
	"context"
	"encoding/hex"
	"fmt"
	"hash"
	"io"

	"github.com/google/uuid"
	"github.com/maklybae/plagiarism-checker/storage/internal/domain"
	"github.com/maklybae/plagiarism-checker/storage/pkg/ctxio"
)

type HashProvider interface {
	Hash() hash.Hash
}

type StorageService struct {
	repo    domain.Repository
	storage domain.Storage
	hash    HashProvider
}

func NewStorageService(repo domain.Repository, storage domain.Storage, hash HashProvider) *StorageService {
	return &StorageService{
		repo:    repo,
		storage: storage,
		hash:    hash,
	}
}

func (s *StorageService) UploadFile(ctx context.Context) (id uuid.UUID, stream io.WriteCloser, errChan <-chan error) {
	id = uuid.New()
	pr, pw := io.Pipe()
	ch := make(chan error, 1)

	go func() {
		defer close(ch)
		defer pr.Close()

		fileWriter, err := s.storage.Save(ctx, id.String())
		if err != nil {
			ch <- fmt.Errorf("failed to upload file: %w", err)
			return
		}
		defer fileWriter.Close()

		hashWriter := s.hash.Hash()

		writer := io.MultiWriter(fileWriter, hashWriter)
		if _, err := io.Copy(writer, ctxio.NewContextReader(ctx, pr)); err != nil {
			ch <- fmt.Errorf("failed to pipe data: %w", err)
			return
		}

		hash := hashWriter.Sum(nil)
		fileInfo := &domain.FileInfo{
			ID:   id,
			Path: id.String(),
			Hash: hex.EncodeToString(hash),
		}

		if err := s.repo.CreateFile(ctx, fileInfo); err != nil {
			ch <- fmt.Errorf("failed to save file info: %w", err)
			return
		}
		ch <- nil
	}()

	return id, pw, ch
}

func (s *StorageService) DownloadFile(ctx context.Context, id uuid.UUID) (stream io.ReadCloser, err error) {
	file, err := s.repo.GetFileByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file to download: %w", err)
	}

	stream, err = s.storage.Get(ctx, file.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file from storage: %w", err)
	}

	return stream, nil
}

func (s *StorageService) GetFileInfo(ctx context.Context, id uuid.UUID) (*domain.FileInfo, error) {
	fileInfo, err := s.repo.GetFileByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return fileInfo, nil
}

func (s *StorageService) ListFilesByHash(ctx context.Context, hash string) ([]*domain.FileInfo, error) {
	if hash == "" {
		return nil, fmt.Errorf("hash cannot be empty")
	}

	files, err := s.repo.GetFilesByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to list files by hash: %w", err)
	}

	return files, nil
}
