package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maklybae/plagiarism-checker/analysis/db"
	"github.com/maklybae/plagiarism-checker/analysis/internal/application"
	"github.com/maklybae/plagiarism-checker/analysis/internal/config"
	"github.com/maklybae/plagiarism-checker/analysis/internal/infrastructure/repository/postgres"
	"github.com/maklybae/plagiarism-checker/analysis/internal/infrastructure/server"
	"github.com/maklybae/plagiarism-checker/analysis/internal/infrastructure/wordcloud"
	pb "github.com/maklybae/plagiarism-checker/genproto/go/analysis"
	storagepb "github.com/maklybae/plagiarism-checker/genproto/go/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	config := config.NewConfig()

	connConfig, err := pgxpool.ParseConfig(config.DSN())
	if err != nil {
		log.Fatal(err)
	}

	db.MustMigrate(connConfig.ConnConfig)

	ctx := context.Background()

	pool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v\n", err)
	}

	repo := postgres.NewRepository(pool)
	wordcloudClient := wordcloud.NewClient()
	wordcloudStore := wordcloud.NewStorage()

	// gRPC-клиент к storage
	storageConn, err := grpc.NewClient(
		config.StorageGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to storage service: %v\n", err)
	}
	defer storageConn.Close()

	storageClient := storagepb.NewStorageServiceClient(storageConn)
	service := application.NewService(repo, wordcloudClient, wordcloudStore, storageClient)

	// Listen all interfaces (debug mode to accept all connections) on port 50051.
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "", config.Port))
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v\n", err)
	}

	grpcServer := grpc.NewServer()
	serverImpl := server.NewAnalysisServer(service)

	pb.RegisterAnalysisServiceServer(grpcServer, serverImpl)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	grpcServer.GracefulStop()

	log.Println("Received shutdown signal, stopping analysis server...")
}
