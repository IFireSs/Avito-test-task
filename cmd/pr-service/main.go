package main

import (
	"avito-autumn2025-internship/internal/app"
	"avito-autumn2025-internship/internal/config"
	"context"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("app stopped with error: %v", err)
	}
}
