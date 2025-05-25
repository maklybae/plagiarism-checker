package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maklybae/plagiarism-checker/analysis/db"
	"github.com/maklybae/plagiarism-checker/analysis/internal/application"
	"github.com/maklybae/plagiarism-checker/analysis/internal/infrastructure/repository/postgres"
	"github.com/maklybae/plagiarism-checker/analysis/internal/infrastructure/server"
	wordcloudclient "github.com/maklybae/plagiarism-checker/analysis/internal/infrastructure/wordcloudclient"
	wordcloudstorage "github.com/maklybae/plagiarism-checker/analysis/internal/infrastructure/wordcloudclient"
	pb "github.com/maklybae/plagiarism-checker/genproto/go/analysis"
	storagepb "github.com/maklybae/plagiarism-checker/genproto/go/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	time.Sleep(5 * time.Second) // Wait for the database to be ready
	dsn := "postgres://postgres:postgres@analysis_db:5432/analysis?sslmode=disable"

	connConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal(err)
	}

	db.MustMigrate(connConfig.ConnConfig)

	ctx := context.Background()

	pool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	repo := postgres.NewRepository(pool)
	wordcloudClient := wordcloudclient.NewClient()
	wordcloudStore := wordcloudstorage.NewStorage()

	// gRPC-клиент к storage
	storageConn, err := grpc.NewClient(
		"storage:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to storage service: %v\n", err)
	}
	defer storageConn.Close()

	storageClient := storagepb.NewStorageServiceClient(storageConn)

	service := application.NewService(repo, wordcloudClient, wordcloudStore, storageClient)

	go func() {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen on port 50051: %v\n", err)
		}

		grpcServer := grpc.NewServer()
		serverImpl := server.NewAnalysisServer(service)

		pb.RegisterAnalysisServiceServer(grpcServer, serverImpl)

		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	log.Println("Received shutdown signal, stopping analysis server...")
}
