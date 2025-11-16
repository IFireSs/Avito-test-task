package app

import (
	"avito-autumn2025-internship/internal/config"
	httptransport "avito-autumn2025-internship/internal/http"
	"avito-autumn2025-internship/internal/repository/postgres"
	"avito-autumn2025-internship/internal/service"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"time"
)

type App struct {
	cfg    config.Config
	server *http.Server
	db     *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	db, err := pgxpool.New(ctx, cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	db.Config().MaxConns = cfg.DB.MaxConns
	db.Config().MaxConnIdleTime = cfg.DB.MaxIdleTime

	teamRepo := postgres.NewTeamRepository(db)
	userRepo := postgres.NewUserRepository(db)
	prRepo := postgres.NewPRRepository(db)

	teamSvc := service.NewTeamService(teamRepo, userRepo)
	userSvc := service.NewUserService(userRepo, prRepo)
	prSvc := service.NewPRService(prRepo, userRepo)

	router := httptransport.NewRouter(prSvc, teamSvc, userSvc, cfg.AdminToken)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		cfg:    cfg,
		server: srv,
		db:     db,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		log.Printf("HTTP server listening on %s", a.cfg.HTTPAddr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Println("shutting down HTTP server...")
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server.Shutdown: %w", err)
		}
		a.db.Close()
		return nil
	case err := <-errCh:
		a.db.Close()
		return err
	}
}
