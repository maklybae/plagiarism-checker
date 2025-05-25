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
	types "github.com/maklybae/plagiarism-checker/genproto/go/storage"
	"github.com/maklybae/plagiarism-checker/storage/db"
	"github.com/maklybae/plagiarism-checker/storage/internal/application"
	"github.com/maklybae/plagiarism-checker/storage/internal/config"
	"github.com/maklybae/plagiarism-checker/storage/internal/infrastructure/hash"
	"github.com/maklybae/plagiarism-checker/storage/internal/infrastructure/repository/postgres"
	"github.com/maklybae/plagiarism-checker/storage/internal/infrastructure/server"
	"github.com/maklybae/plagiarism-checker/storage/internal/infrastructure/storage/local"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.NewConfig()
	connConfig, err := pgxpool.ParseConfig(cfg.DSN())
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
	service := application.NewStorageService(repo, local.NewStorage(), hash.NewSHA256())

	grpcServer := grpc.NewServer()
	serverImpl := server.NewStorageServer(service)

	types.RegisterStorageServiceServer(grpcServer, serverImpl)

	// Listen all interfaces (debug mode to accept all connections) on port 50051.
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "", cfg.Port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v\n", cfg.Port, err)
	}

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Waiting for SIGINT (pkill -2) or SIGTERM
	<-stop

	// Graceful shutdown
	grpcServer.GracefulStop()

	log.Println("Received shutdown signal, stopping server...")
}
