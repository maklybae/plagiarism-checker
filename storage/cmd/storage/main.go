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
	types "github.com/maklybae/plagiarism-checker/genproto/go/storage"
	"github.com/maklybae/plagiarism-checker/storage/db"
	"github.com/maklybae/plagiarism-checker/storage/internal/application"
	"github.com/maklybae/plagiarism-checker/storage/internal/infrastructure/hash"
	"github.com/maklybae/plagiarism-checker/storage/internal/infrastructure/repository/postgres"
	"github.com/maklybae/plagiarism-checker/storage/internal/infrastructure/server"
	"github.com/maklybae/plagiarism-checker/storage/internal/infrastructure/storage/local"
	"google.golang.org/grpc"
)

func main() {
	time.Sleep(5 * time.Second) // Wait for the database to be ready
	dsn := "postgres://postgres:postgres@storage_db:5432/storage?sslmode=disable"

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
	service := application.NewStorageService(repo, local.NewStorage(), hash.NewSHA256())

	go func() {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen on port 50051: %v\n", err)
		}

		grpcServer := grpc.NewServer()
		serverImpl := server.NewStorageServer(service)

		types.RegisterStorageServiceServer(grpcServer, serverImpl)

		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Waiting for SIGINT (pkill -2) or SIGTERM
	<-stop
	log.Println("Received shutdown signal, stopping server...")
}
