package server

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/maklybae/plagiarism-checker/analysis/internal/application"
	pb "github.com/maklybae/plagiarism-checker/genproto/go/analysis"
	"github.com/maklybae/plagiarism-checker/genproto/go/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AnalysisServer struct {
	pb.UnimplementedAnalysisServiceServer
	service *application.Service
}

func NewAnalysisServer(service *application.Service) *AnalysisServer {
	return &AnalysisServer{service: service}
}

func safeIntToUint64(val int) uint64 {
	if val < 0 {
		return 0
	}

	return uint64(val)
}

func (s *AnalysisServer) SingleFileAnalysis(
	ctx context.Context,
	req *pb.SingleFileAnalysisRequest,
) (*pb.SingleFileAnalysisResponse, error) {
	fileID, err := uuid.Parse(req.GetFileId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid file_id: %v", err)
	}

	textInfo, exactMatch, err := s.service.SingleFileAnalysis(ctx, fileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "analysis error: %v", err)
	}

	if textInfo == nil {
		return nil, status.Error(codes.NotFound, "no analysis result")
	}

	dtoExactMatch := make([]*common.UUID, 0, len(exactMatch))
	for _, id := range exactMatch {
		dtoExactMatch = append(dtoExactMatch, &common.UUID{Value: id.String()})
	}

	resp := &pb.SingleFileAnalysisResponse{
		FileId: &common.UUID{Value: fileID.String()},
		Statistics: &pb.FileStatistics{
			TotalLines:      safeIntToUint64(textInfo.Lines),
			TotalWords:      safeIntToUint64(textInfo.Words),
			TotalCharacters: safeIntToUint64(textInfo.Chars),
		},
		ExactMatches: dtoExactMatch,
	}

	return resp, nil
}

func (s *AnalysisServer) DiffFileAnalysis(ctx context.Context, req *pb.DiffFileAnalysisRequest) (*pb.DiffFileAnalysisResponse, error) {
	fileA, err := uuid.Parse(req.GetFileIdA().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid file_id_a: %v", err)
	}

	fileB, err := uuid.Parse(req.GetFileIdB().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid file_id_b: %v", err)
	}

	sim, err := s.service.DiffFileAnalysis(ctx, fileA, fileB)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "diff analysis error: %v", err)
	}

	resp := &pb.DiffFileAnalysisResponse{
		FileIdA:    &common.UUID{Value: fileA.String()},
		FileIdB:    &common.UUID{Value: fileB.String()},
		Similarity: sim,
	}

	return resp, nil
}

func (s *AnalysisServer) WordCloud(ctx context.Context, req *pb.WordCloudRequest) (*pb.WordCloudResponse, error) {
	fileID, err := uuid.Parse(req.GetFileId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid file_id: %v", err)
	}

	imgReader, err := s.service.WordCloud(ctx, fileID)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Error(codes.NotFound, "file not found")
		}

		return nil, status.Errorf(codes.Internal, "wordcloud error: %v", err)
	}
	defer imgReader.Close()

	imgBytes, err := io.ReadAll(imgReader)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read wordcloud image: %v", err)
	}

	resp := &pb.WordCloudResponse{
		File: &common.FileChunk{Content: imgBytes},
	}

	return resp, nil
}
