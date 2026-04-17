package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port       int
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string
	JWTSecret  string

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioSecure    bool
}

func Load() *Config {
	return &Config{
		Port:       envInt("API_PORT", 8080),
		DBHost:     env("POSTGRES_HOST", "localhost"),
		DBPort:     envInt("POSTGRES_PORT", 5432),
		DBName:     env("POSTGRES_DB", "brandhunt"),
		DBUser:     env("POSTGRES_USER", "brandhunt"),
		DBPassword: env("POSTGRES_PASSWORD", "changeme"),
		JWTSecret:  env("JWT_SECRET", "brandhunt-secret-change-me"),

		MinioEndpoint:  env("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: env("MINIO_ROOT_USER", "minioadmin"),
		MinioSecretKey: env("MINIO_ROOT_PASSWORD", "changeme123"),
		MinioBucket:    env("MINIO_BUCKET", "brandhunt-photos"),
		MinioSecure:    env("MINIO_SECURE", "false") == "true",
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBName, c.DBUser, c.DBPassword,
	)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
