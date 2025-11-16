package postgres

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/repository"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) UpsertTeamMembers(
	ctx context.Context,
	teamName string,
	members []api.TeamMember,
) ([]api.User, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	users := make([]api.User, 0, len(members))

	for _, m := range members {
		var u api.User
		err := tx.QueryRow(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) DO UPDATE
			    SET username = EXCLUDED.username,
			        team_name = EXCLUDED.team_name,
			        is_active = EXCLUDED.is_active
			RETURNING user_id, username, team_name, is_active
		`,
			m.UserId,
			m.Username,
			teamName,
			m.IsActive,
		).Scan(&u.UserId, &u.Username, &u.TeamName, &u.IsActive)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) ListByTeam(ctx context.Context, teamName string) ([]api.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY user_id
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []api.User
	for rows.Next() {
		var u api.User
		if err := rows.Scan(&u.UserId, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return users, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID string) (*api.User, error) {
	var u api.User
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`, userID).Scan(&u.UserId, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {
	var u api.User
	err := r.pool.QueryRow(ctx, `
		UPDATE users
		SET is_active = $2
		WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active
	`, userID, isActive).Scan(&u.UserId, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) ListActiveByTeam(ctx context.Context, teamName string) ([]api.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 AND is_active = TRUE
		ORDER BY user_id
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []api.User
	for rows.Next() {
		var user api.User
		if err := rows.Scan(&user.UserId, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return users, nil
}
