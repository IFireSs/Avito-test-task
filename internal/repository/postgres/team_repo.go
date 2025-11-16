package postgres

import (
	"avito-autumn2025-internship/internal/repository"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type teamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) repository.TeamRepository {
	return &teamRepository{pool: pool}
}

func (r *teamRepository) Create(ctx context.Context, teamName string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO teams (team_name) 
		VALUES ($1)
	`, teamName)
	return err
}

func (r *teamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)
	`, teamName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
