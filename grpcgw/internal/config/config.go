package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port             int    `env:"PORT" envDefault:"8080"`
	AnalysisGRPCAddr string `env:"ANALYSIS_GRPC_ADDR" envDefault:"analysis:50051"`
	StorageGRPCAddr  string `env:"STORAGE_GRPC_ADDR" envDefault:"storage:50051"`
}

func NewConfig() *Config {
	cfg := env.Must(env.ParseAs[Config]())
	return &cfg
}

func (c *Config) ListenAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}
