package application_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/maklybae/plagiarism-checker/storage/internal/application"
	mocksapp "github.com/maklybae/plagiarism-checker/storage/internal/application/mocks"
	"github.com/maklybae/plagiarism-checker/storage/internal/domain"
	mocksdomain "github.com/maklybae/plagiarism-checker/storage/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type nopWriteCloser struct {
	*bytes.Buffer
}

func (n *nopWriteCloser) Close() error { return nil }

func TestStorageService_GetFileInfo(t *testing.T) {
	repo := new(mocksdomain.Repository)
	storage := new(mocksdomain.Storage)
	hash := new(mocksapp.HashProvider)
	service := application.NewStorageService(repo, storage, hash)

	ctx := context.Background()
	fileID := uuid.New()
	fileInfo := &domain.FileInfo{ID: fileID, Path: "somepath", Hash: "abc"}

	t.Run("success", func(t *testing.T) {
		repo.On("GetFileByID", ctx, fileID).Return(fileInfo, nil).Once()
		result, err := service.GetFileInfo(ctx, fileID)
		require.NoError(t, err)
		assert.Equal(t, fileInfo, result)
		repo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		repo.On("GetFileByID", ctx, fileID).Return(nil, errors.New("not found")).Once()
		result, err := service.GetFileInfo(ctx, fileID)
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to get file info")
		repo.AssertExpectations(t)
	})
}

func TestStorageService_ListFilesByHash(t *testing.T) {
	repo := new(mocksdomain.Repository)
	storage := new(mocksdomain.Storage)
	hash := new(mocksapp.HashProvider)
	service := application.NewStorageService(repo, storage, hash)

	ctx := context.Background()
	hashStr := "abc123"
	files := []*domain.FileInfo{{ID: uuid.New(), Path: "p", Hash: hashStr}}

	t.Run("success", func(t *testing.T) {
		repo.On("GetFilesByHash", ctx, hashStr).Return(files, nil).Once()
		result, err := service.ListFilesByHash(ctx, hashStr)
		require.NoError(t, err)
		assert.Equal(t, files, result)
		repo.AssertExpectations(t)
	})

	t.Run("empty hash", func(t *testing.T) {
		result, err := service.ListFilesByHash(ctx, "")
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "hash cannot be empty")
	})

	t.Run("repo error", func(t *testing.T) {
		repo.On("GetFilesByHash", ctx, hashStr).Return(nil, errors.New("db error")).Once()
		result, err := service.ListFilesByHash(ctx, hashStr)
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to list files by hash")
		repo.AssertExpectations(t)
	})
}

func TestStorageService_DownloadFile(t *testing.T) {
	repo := new(mocksdomain.Repository)
	storage := new(mocksdomain.Storage)
	hash := new(mocksapp.HashProvider)
	service := application.NewStorageService(repo, storage, hash)

	ctx := context.Background()
	fileID := uuid.New()
	fileInfo := &domain.FileInfo{ID: fileID, Path: "somepath", Hash: "abc"}
	fileContent := "testdata"
	readCloser := io.NopCloser(strings.NewReader(fileContent))

	t.Run("success", func(t *testing.T) {
		repo.On("GetFileByID", ctx, fileID).Return(fileInfo, nil).Once()
		storage.On("Get", ctx, fileInfo.Path).Return(readCloser, nil).Once()
		result, err := service.DownloadFile(ctx, fileID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		data, err := io.ReadAll(result)
		require.NoError(t, err)
		assert.Equal(t, fileContent, string(data))
		repo.AssertExpectations(t)
		storage.AssertExpectations(t)
	})

	t.Run("file not found", func(t *testing.T) {
		repo.On("GetFileByID", ctx, fileID).Return(nil, errors.New("not found")).Once()
		result, err := service.DownloadFile(ctx, fileID)
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to get file to download")
		repo.AssertExpectations(t)
	})

	t.Run("storage error", func(t *testing.T) {
		repo.On("GetFileByID", ctx, fileID).Return(fileInfo, nil).Once()
		storage.On("Get", ctx, fileInfo.Path).Return(nil, errors.New("storage error")).Once()
		result, err := service.DownloadFile(ctx, fileID)
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to get file from storage")
		repo.AssertExpectations(t)
		storage.AssertExpectations(t)
	})
}

func TestStorageService_UploadFile(t *testing.T) {
	ctx := context.Background()
	fileContent := "testdata123"

	t.Run("success", func(t *testing.T) {
		repo := new(mocksdomain.Repository)
		storage := new(mocksdomain.Storage)
		hashProvider := new(mocksapp.HashProvider)
		service := application.NewStorageService(repo, storage, hashProvider)

		fileWriter := &bytes.Buffer{}
		fileWriterCloser := &nopWriteCloser{fileWriter}
		hashWriter := sha256.New()
		hashProvider.On("Hash").Return(hashWriter).Maybe()
		storage.On("Save", ctx, mock.Anything).Return(fileWriterCloser, nil).Once()
		repo.On("CreateFile", ctx, mock.MatchedBy(func(info *domain.FileInfo) bool {
			expectedHash := sha256.Sum256([]byte(fileContent))
			return info.Hash == hex.EncodeToString(expectedHash[:])
		})).Return(nil).Once()

		_, writer, errCh := service.UploadFile(ctx)
		_, err := writer.Write([]byte(fileContent))
		assert.NoError(t, err)
		writer.Close()
		err = <-errCh
		assert.NoError(t, err)
		repo.AssertExpectations(t)
		storage.AssertExpectations(t)
	})

	t.Run("storage error", func(t *testing.T) {
		repo := new(mocksdomain.Repository)
		storage := new(mocksdomain.Storage)
		hashProvider := new(mocksapp.HashProvider)
		service := application.NewStorageService(repo, storage, hashProvider)
		hashWriter := sha256.New()
		hashProvider.On("Hash").Return(hashWriter).Maybe()

		storage.On("Save", ctx, mock.Anything).Return(nil, assert.AnError).Once()
		id, writer, errCh := service.UploadFile(ctx)
		assert.NotEqual(t, uuid.Nil, id)
		assert.NotNil(t, writer)
		writer.Close()
		err := <-errCh
		assert.ErrorContains(t, err, "failed to upload file")
		storage.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := new(mocksdomain.Repository)
		storage := new(mocksdomain.Storage)
		hashProvider := new(mocksapp.HashProvider)
		service := application.NewStorageService(repo, storage, hashProvider)
		hashWriter := sha256.New()
		hashProvider.On("Hash").Return(hashWriter).Maybe()

		storage.On("Save", ctx, mock.Anything).Return(&nopWriteCloser{&bytes.Buffer{}}, nil).Once()
		repo.On("CreateFile", ctx, mock.Anything).Return(assert.AnError).Once()
		_, writer, errCh := service.UploadFile(ctx)
		_, _ = writer.Write([]byte(fileContent))
		writer.Close()
		err := <-errCh
		assert.ErrorContains(t, err, "failed to save file info")
		storage.AssertExpectations(t)
		repo.AssertExpectations(t)
	})
}
