package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maklybae/plagiarism-checker/genproto/go/analysis"
	"github.com/maklybae/plagiarism-checker/genproto/go/common"
	"github.com/maklybae/plagiarism-checker/genproto/go/storage"
	"github.com/maklybae/plagiarism-checker/grpcgw/internal/types"
	"google.golang.org/grpc/status"
)

// Handler holds gRPC clients for analysis and storage services.
type Handler struct {
	AnalysisClient analysis.AnalysisServiceClient
	StorageClient  storage.StorageServiceClient
}

// Ensure Handler implements types.ServerInterface
var _ types.ServerInterface = (*Handler)(nil)

// Analyze a single file
func (h *Handler) PostApiV1AnalysisSingle(c *gin.Context) {
	var req types.PostApiV1AnalysisSingleJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.Error{Code: intPtr(http.StatusBadRequest), Message: strPtr("invalid request body")})
		return
	}
	resp, err := h.AnalysisClient.SingleFileAnalysis(c.Request.Context(), &analysis.SingleFileAnalysisRequest{
		FileId: &common.UUID{Value: req.FileId.String()},
	})
	if err != nil {
		handleGrpcError(c, err)
		return
	}
	matches := make([]string, 0, len(resp.ExactMatches))
	for _, id := range resp.ExactMatches {
		matches = append(matches, id.GetValue())
	}
	c.JSON(http.StatusOK, gin.H{
		"file_id": resp.GetFileId().GetValue(),
		"statistics": gin.H{
			"total_lines":      resp.GetStatistics().GetTotalLines(),
			"total_words":      resp.GetStatistics().GetTotalWords(),
			"total_characters": resp.GetStatistics().GetTotalCharacters(),
		},
		"exact_matches": matches,
	})
}

// Compare two files
func (h *Handler) PostApiV1AnalysisDiff(c *gin.Context) {
	var req types.PostApiV1AnalysisDiffJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.Error{Code: intPtr(http.StatusBadRequest), Message: strPtr("invalid request body")})
		return
	}
	resp, err := h.AnalysisClient.DiffFileAnalysis(c.Request.Context(), &analysis.DiffFileAnalysisRequest{
		FileIdA: &common.UUID{Value: req.FileIdA.String()},
		FileIdB: &common.UUID{Value: req.FileIdB.String()},
	})
	if err != nil {
		handleGrpcError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"file_id_a":  resp.GetFileIdA().GetValue(),
		"file_id_b":  resp.GetFileIdB().GetValue(),
		"similarity": resp.GetSimilarity(),
	})
}

// Generate wordcloud for a file
func (h *Handler) PostApiV1AnalysisWordcloud(c *gin.Context) {
	var req types.PostApiV1AnalysisWordcloudJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.Error{Code: intPtr(http.StatusBadRequest), Message: strPtr("invalid request body")})
		return
	}
	resp, err := h.AnalysisClient.WordCloud(c.Request.Context(), &analysis.WordCloudRequest{
		FileId: &common.UUID{Value: req.FileId.String()},
	})
	if err != nil {
		handleGrpcError(c, err)
		return
	}
	if resp.File == nil || len(resp.File.Content) == 0 {
		c.JSON(http.StatusNotFound, types.Error{Code: intPtr(http.StatusNotFound), Message: strPtr("wordcloud not found")})
		return
	}
	c.Data(http.StatusOK, "image/png", resp.File.Content)
}

// Download a file
func (h *Handler) GetApiV1Storage(c *gin.Context, params types.GetApiV1StorageParams) {
	respStream, err := h.StorageClient.DownloadFile(c.Request.Context(), &storage.DownloadFileRequest{
		FileId: &common.UUID{Value: params.FileId.String()},
	})
	if err != nil {
		handleGrpcError(c, err)
		return
	}
	var fileContent []byte
	for {
		resp, err := respStream.Recv()
		if err != nil {
			break
		}
		if resp.File != nil {
			fileContent = append(fileContent, resp.File.Content...)
		}
	}
	if len(fileContent) == 0 {
		c.JSON(http.StatusNotFound, types.Error{Code: intPtr(http.StatusNotFound), Message: strPtr("file not found")})
		return
	}
	c.Data(http.StatusOK, "application/octet-stream", fileContent)
}

// Upload a file
func (h *Handler) PostApiV1Storage(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil || len(data) == 0 {
		c.JSON(http.StatusBadRequest, types.Error{Code: intPtr(http.StatusBadRequest), Message: strPtr("empty or invalid file upload")})
		return
	}
	stream, err := h.StorageClient.UploadFile(c.Request.Context())
	if err != nil {
		handleGrpcError(c, err)
		return
	}
	err = stream.Send(&storage.UploadFileRequest{
		File: &common.FileChunk{Content: data},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.Error{Code: intPtr(http.StatusInternalServerError), Message: strPtr("failed to send file")})
		return
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		handleGrpcError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"file_id": resp.GetFileId().GetValue()})
}

// Helper: handle gRPC error and map to HTTP
func handleGrpcError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, types.Error{Code: intPtr(http.StatusInternalServerError), Message: strPtr("internal error")})
		return
	}
	switch st.Code() {
	case 3:
		c.JSON(http.StatusBadRequest, types.Error{Code: intPtr(http.StatusBadRequest), Message: strPtr(st.Message())})
	case 5:
		c.JSON(http.StatusNotFound, types.Error{Code: intPtr(http.StatusNotFound), Message: strPtr(st.Message())})
	default:
		c.JSON(http.StatusInternalServerError, types.Error{Code: intPtr(http.StatusInternalServerError), Message: strPtr(st.Message())})
	}
}

func intPtr(i int) *int       { return &i }
func strPtr(s string) *string { return &s }
