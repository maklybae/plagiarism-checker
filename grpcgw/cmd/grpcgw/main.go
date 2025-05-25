package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maklybae/plagiarism-checker/genproto/go/analysis"
	"github.com/maklybae/plagiarism-checker/genproto/go/storage"
	"github.com/maklybae/plagiarism-checker/grpcgw/internal/config"
	"github.com/maklybae/plagiarism-checker/grpcgw/internal/server"
	"github.com/maklybae/plagiarism-checker/grpcgw/internal/types"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.NewConfig()

	// gRPC clients setup
	analysisConn, err := grpc.NewClient(
		cfg.AnalysisGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to analysis: %v", err)
	}
	defer analysisConn.Close()

	storageConn, err := grpc.NewClient(
		cfg.StorageGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to storage: %v", err)
	}
	defer storageConn.Close()

	h := &server.Handler{
		AnalysisClient: analysis.NewAnalysisServiceClient(analysisConn),
		StorageClient:  storage.NewStorageServiceClient(storageConn),
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// OpenAPI documentation and Swagger UI setup
	oapiPath := filepath.Join("..", "openapi", "v1", "api.yaml")
	r.StaticFile("/swagger.yaml", oapiPath)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger.yaml")))

	types.RegisterHandlersWithOptions(r, h, types.GinServerOptions{
		BaseURL: "",
		ErrorHandler: func(c *gin.Context, err error, code int) {
			c.JSON(code, types.Error{Code: &code, Message: strPtr(err.Error())})
		},
	})

	srv := &http.Server{
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         cfg.ListenAddr(),
	}

	go func() {
		log.Printf("HTTP server listening on %s", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func strPtr(s string) *string { return &s }
