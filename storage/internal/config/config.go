package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port       int    `env:"PORT" envDefault:"50051"`
	DBHost     string `env:"DB_HOST" envDefault:"storage_db"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER" envDefault:"postgres"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"postgres"`
	DBName     string `env:"DB_NAME" envDefault:"storage"`
}

func NewConfig() *Config {
	cfg := env.Must(env.ParseAs[Config]())
	return &cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}
