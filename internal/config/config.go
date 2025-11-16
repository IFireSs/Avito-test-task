package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type DBConfig struct {
	DSN         string
	MaxConns    int32
	MaxIdleTime time.Duration
}

type Config struct {
	HTTPAddr   string
	AdminToken string
	DB         DBConfig
}

func MustLoad() Config {
	cfg := Config{}

	cfg.HTTPAddr = getenv("HTTP_ADDR", ":8080")
	cfg.AdminToken = os.Getenv("ADMIN_TOKEN")

	cfg.DB = DBConfig{
		DSN:         getenv("DB_DSN", "postgres://postgres:postgres@db:5432/postgres?sslmode=disable"),
		MaxConns:    getInt32("DB_MAX_CONNS", 10),
		MaxIdleTime: getDuration("DB_MAX_IDLE_TIME", "1m"),
	}

	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt32(key string, def int32) int32 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("invalid %s value %q: %v", key, v, err)
	}
	return int32(n)
}

func getDuration(key, def string) time.Duration {
	v := getenv(key, def)
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Fatalf("invalid %s duration %q: %v", key, v, err)
	}
	return d
}
