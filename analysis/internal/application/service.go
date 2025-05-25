package application

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/maklybae/plagiarism-checker/analysis/internal/domain"
	"github.com/maklybae/plagiarism-checker/analysis/pkg/textstat"
	"github.com/maklybae/plagiarism-checker/genproto/go/common"
	"github.com/maklybae/plagiarism-checker/genproto/go/storage"
)

type WordcloudClient interface {
	GetImage(ctx context.Context, text string) (reader io.ReadCloser, err error)
}

type Service struct {
	repo            domain.Repository
	wordcloudClient WordcloudClient
	wordcloudStore  domain.WordcloudStorage
	storageClient   storage.StorageServiceClient
}

func NewService(
	repo domain.Repository,
	wcClient WordcloudClient,
	wcStore domain.WordcloudStorage,
	storageClient storage.StorageServiceClient,
) *Service {
	return &Service{
		repo:            repo,
		wordcloudClient: wcClient,
		wordcloudStore:  wcStore,
		storageClient:   storageClient,
	}
}

func (s *Service) SingleFileAnalysis(
	ctx context.Context,
	fileID uuid.UUID,
) (textInfo *domain.TextInfo, exactMatchIDs []uuid.UUID, err error) {
	// Check if text info already exists for this file
	var zero *FileInfoNotFoundError

	textInfo, err = s.repo.GetTextInfoByID(ctx, fileID)
	if errors.As(err, &zero) {
		// If we got a "not found" error, we proceed to create text info
		content, err := s.getFileContent(ctx, fileID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get file content: %w", err)
		}

		lines, words, chars := textstat.CountStats(content)
		textInfo = &domain.TextInfo{
			ID:    fileID,
			Lines: lines,
			Words: words,
			Chars: chars,
		}

		if err = s.repo.CreateTextInfo(ctx, textInfo); err != nil {
			return nil, nil, fmt.Errorf("failed to create text info: %w", err)
		}
	} else if err != nil {
		// If we got an error other than "not found", return it
		return nil, nil, fmt.Errorf("failed to get text info: %w", err)
	}

	// Get current file hash
	hashResp, err := s.storageClient.GetFileHash(ctx, &storage.GetFileHashRequest{
		FileId: &common.UUID{Value: fileID.String()},
	})
	if err != nil {
		return textInfo, nil, fmt.Errorf("failed to get file hash: %w", err)
	}

	// List all files with the same hash
	listResp, err := s.storageClient.ListFilesByHash(ctx, &storage.ListFilesByHashRequest{
		Hash: hashResp.GetHash(),
	})
	if err != nil {
		return textInfo, nil, fmt.Errorf("failed to list files by hash: %w", err)
	}

	var ids []uuid.UUID

	for _, pbID := range listResp.GetFileIds() {
		id, err := uuid.Parse(pbID.GetValue())
		if err == nil && id != fileID {
			ids = append(ids, id)
		}
	}

	return textInfo, ids, nil
}

// User should call SingleFileAnalysis before DiffFileAnalysis.
func (s *Service) DiffFileAnalysis(ctx context.Context, fileA, fileB uuid.UUID) (similarity float64, err error) {
	contentA, err := s.getFileContent(ctx, fileA)
	if err != nil {
		return 0, fmt.Errorf("failed to get content for file A: %w", err)
	}

	contentB, err := s.getFileContent(ctx, fileB)
	if err != nil {
		return 0, fmt.Errorf("failed to get content for file B: %w", err)
	}

	sim := textstat.Similarity(string(contentA), string(contentB))

	return sim, nil
}

// User should call SingleFileAnalysis before WordCloud.
func (s *Service) WordCloud(ctx context.Context, fileID uuid.UUID) (stream io.ReadCloser, err error) {
	// Check if wordcloud already exists for this file
	textInfo, err := s.repo.GetTextInfoByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get text info: %w", err)
	}

	if textInfo.WordcloudPath != "" {
		// Check if wordcloud image already exists
		reader, err := s.wordcloudStore.Get(ctx, textInfo.WordcloudPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing wordcloud image: %w", err)
		}

		return reader, nil
	}

	content, err := s.getFileContent(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content for wordcloud: %w", err)
	}

	imgReader, err := s.wordcloudClient.GetImage(ctx, string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to get wordcloud image: %w", err)
	}
	defer imgReader.Close()

	path := fmt.Sprintf("%s.png", fileID.String())

	imgWriter, err := s.wordcloudStore.Save(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to save wordcloud image: %w", err)
	}

	_, err = io.Copy(imgWriter, imgReader)
	if err != nil {
		return nil, fmt.Errorf("failed to write wordcloud image: %w", err)
	}

	imgWriter.Close()

	if err = s.repo.SetWordcloudPath(ctx, fileID, path); err != nil {
		return nil, fmt.Errorf("failed to update text info with wordcloud path: %w", err)
	}

	reader, err := s.wordcloudStore.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to open saved wordcloud image: %w", err)
	}

	return reader, nil
}

func (s *Service) getFileContent(ctx context.Context, fileID uuid.UUID) ([]byte, error) {
	inStream, err := s.storageClient.DownloadFile(ctx, &storage.DownloadFileRequest{
		FileId: &common.UUID{Value: fileID.String()},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	var content []byte

	for {
		resp, err := inStream.Recv()
		if err != nil {
			if err == io.EOF {
				if err := inStream.CloseSend(); err != nil {
					return nil, fmt.Errorf("failed to close stream: %w", err)
				}

				break
			}

			return nil, err
		}

		if resp.File != nil {
			content = append(content, resp.File.Content...)
		}
	}

	return content, nil
}
