package server

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/maklybae/plagiarism-checker/genproto/go/common"
	types "github.com/maklybae/plagiarism-checker/genproto/go/storage"
	"github.com/maklybae/plagiarism-checker/storage/internal/application"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StorageServer struct {
	types.UnimplementedStorageServiceServer
	service *application.StorageService
}

func NewStorageServer(service *application.StorageService) *StorageServer {
	return &StorageServer{
		service: service,
	}
}

func (s *StorageServer) UploadFile(serverStream grpc.ClientStreamingServer[types.UploadFileRequest, types.UploadFileResponse]) error {
	id, stream, errChan := s.service.UploadFile(serverStream.Context())

	for {
		select {
		case <-serverStream.Context().Done():
			return status.Error(codes.Canceled, "upload canceled by client")
		default:
		}

		req, err := serverStream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// End of stream, finalize the upload
				if err := stream.Close(); err != nil {
					return status.Errorf(codes.Internal, "failed to close stream: %v", err)
				}

				break
			}

			return status.Errorf(codes.Internal, "failed to receive request: %v", err)
		}

		if _, err = stream.Write(req.GetFile().GetContent()); err != nil {
			return status.Errorf(codes.Internal, "failed to write to stream: %v", err)
		}
	}

	// Wait until UploadFile errChan is done
	if err := <-errChan; err != nil {
		return status.Errorf(codes.Internal, "failed to upload file: %v", err)
	}

	return serverStream.SendAndClose(&types.UploadFileResponse{
		FileId: &common.UUID{Value: id.String()},
	})
}

func (s *StorageServer) DownloadFile(
	req *types.DownloadFileRequest,
	serverStream grpc.ServerStreamingServer[types.DownloadFileResponse],
) error {
	bufferSize := 64 * 1024 // 64 KB

	id, err := uuid.Parse(req.GetFileId().GetValue())
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid file ID: %v", err)
	}

	stream, err := s.service.DownloadFile(serverStream.Context(), id)
	if err != nil {
		var zero *application.FileNotFoundError
		if errors.As(err, &zero) {
			return status.Error(codes.NotFound, "file not found")
		}

		return status.Errorf(codes.Internal, "failed to download file: %v", err)
	}

	buff := make([]byte, bufferSize)

	for {
		bytesRead, err := stream.Read(buff)
		if err != nil {
			if errors.Is(err, io.EOF) {
				// End of stream, break the loop
				break
			}

			return status.Errorf(codes.Internal, "failed to read from stream: %v", err)
		}

		resp := &types.DownloadFileResponse{
			File: &common.FileChunk{Content: buff[:bytesRead]},
		}
		if err := serverStream.Send(resp); err != nil {
			return status.Errorf(codes.Internal, "failed to send response: %v", err)
		}
	}

	if err := stream.Close(); err != nil {
		return status.Errorf(codes.Internal, "failed to close stream: %v", err)
	}

	return nil
}

func (s *StorageServer) GetFileHash(ctx context.Context, req *types.GetFileHashRequest) (*types.GetFileHashResponse, error) {
	id, err := uuid.Parse(req.GetFileId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid file ID: %v", err)
	}

	hash, err := s.service.GetFileInfo(ctx, id)
	if err != nil {
		var zero *application.FileNotFoundError
		if errors.As(err, &zero) {
			return nil, status.Error(codes.NotFound, "file not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to get file hash: %v", err)
	}

	return &types.GetFileHashResponse{
		FileId: &common.UUID{Value: id.String()},
		Hash:   hash.Hash,
	}, nil
}

func (s *StorageServer) ListFilesByHash(ctx context.Context, req *types.ListFilesByHashRequest) (*types.ListFilesByHashResponse, error) {
	hash := req.GetHash()
	if hash == "" {
		return nil, status.Error(codes.InvalidArgument, "hash cannot be empty")
	}

	files, err := s.service.ListFilesByHash(ctx, hash)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list files by hash: %v", err)
	}

	ids := make([]*common.UUID, 0, len(files))
	for _, file := range files {
		ids = append(ids, &common.UUID{
			Value: file.ID.String(),
		})
	}

	return &types.ListFilesByHashResponse{
		FileIds: ids,
	}, nil
}
