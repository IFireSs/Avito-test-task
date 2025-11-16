package tests

import (
	"avito-autumn2025-internship/internal/api"
	pgrepo "avito-autumn2025-internship/internal/repository/postgres"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func connectTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://pr_test_user:pr_test_password@localhost:5433/pr_db_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("skip postgres repo tests: cannot create pool for %q: %v", dsn, err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skip postgres repo tests: cannot ping %q: %v", dsn, err)
	}

	t.Cleanup(func() { pool.Close() })

	return pool
}

func truncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx,
		`TRUNCATE TABLE 
		    pull_request_reviewers,
		    pull_requests,
		    users,
		    teams
		 RESTART IDENTITY CASCADE`,
	)
	require.NoError(t, err)
}

func TestPostgresUserRepository_UpsertAndList(t *testing.T) {
	pool := connectTestDB(t)
	truncateAll(t, pool)

	ctx := context.Background()

	userRepo := pgrepo.NewUserRepository(pool)

	_, err := pool.Exec(ctx,
		"INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING",
		"backend",
	)
	require.NoError(t, err)

	members := []api.TeamMember{
		{UserId: "u1", Username: "dev1", IsActive: true},
		{UserId: "u2", Username: "dev2", IsActive: true},
	}

	users, err := userRepo.UpsertTeamMembers(ctx, "backend", members)
	require.NoError(t, err)
	require.Len(t, users, 2)

	list, err := userRepo.ListByTeam(ctx, "backend")
	require.NoError(t, err)
	require.Len(t, list, 2)

	_, err = userRepo.SetIsActive(ctx, "u2", false)
	require.NoError(t, err)

	u2, err := userRepo.GetByID(ctx, "u2")
	require.NoError(t, err)
	require.NotNil(t, u2)
	require.False(t, u2.IsActive)

	active, err := userRepo.ListActiveByTeam(ctx, "backend")
	require.NoError(t, err)
	require.Len(t, active, 1)
	require.Equal(t, "u1", active[0].UserId)
}

func TestPostgresPRRepository_CreateReplaceAndListShort(t *testing.T) {
	pool := connectTestDB(t)
	truncateAll(t, pool)

	ctx := context.Background()

	userRepo := pgrepo.NewUserRepository(pool)
	prRepo := pgrepo.NewPRRepository(pool)

	_, err := pool.Exec(ctx,
		"INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING",
		"backend",
	)
	require.NoError(t, err)

	_, err = userRepo.UpsertTeamMembers(ctx, "backend", []api.TeamMember{
		{UserId: "u_author", Username: "author"},
		{UserId: "u_old", Username: "old_reviewer"},
		{UserId: "u_new", Username: "new_reviewer"},
		{UserId: "u_dev1", Username: "dev1"},
	})
	require.NoError(t, err)
	now := time.Now().UTC()
	prReassign := &api.PullRequest{
		PullRequestId:     "pr-reassign",
		PullRequestName:   "Reassign test",
		AuthorId:          "u_author",
		Status:            api.PullRequestStatusOPEN,
		AssignedReviewers: []string{"u_old"},
		CreatedAt:         &now,
	}
	err = prRepo.Create(ctx, prReassign)
	require.NoError(t, err)

	prOpen := &api.PullRequest{
		PullRequestId:     "pr-open",
		PullRequestName:   "Open PR",
		AuthorId:          "u_author",
		Status:            api.PullRequestStatusOPEN,
		AssignedReviewers: []string{"u_dev1"},
		CreatedAt:         &now,
	}

	prMerged := &api.PullRequest{
		PullRequestId:     "pr-merged",
		PullRequestName:   "Merged PR",
		AuthorId:          "u_author",
		Status:            api.PullRequestStatusOPEN,
		AssignedReviewers: []string{"u_dev1"},
		CreatedAt:         &now,
	}

	require.NoError(t, prRepo.Create(ctx, prOpen))
	require.NoError(t, prRepo.Create(ctx, prMerged))

	_, err = prRepo.SetMerged(ctx, "pr-merged", time.Now().UTC())
	require.NoError(t, err)

	err = prRepo.ReplaceReviewer(ctx, "pr-reassign", "u_old", "u_new")
	require.NoError(t, err)

	stored, err := prRepo.GetByID(ctx, "pr-reassign")
	require.NoError(t, err)
	require.NotNil(t, stored)
	require.Len(t, stored.AssignedReviewers, 1)
	require.Equal(t, "u_new", stored.AssignedReviewers[0])

	shorts, err := prRepo.ListShortByReviewer(ctx, "u_dev1")
	require.NoError(t, err)
	require.Len(t, shorts, 2, "ожидаем 2 PR для ревьюера u_dev1")

	statuses := make([]api.PullRequestShortStatus, 0, len(shorts))
	for _, s := range shorts {
		statuses = append(statuses, s.Status)
	}

	require.ElementsMatch(t,
		[]api.PullRequestShortStatus{
			api.PullRequestShortStatusOPEN,
			api.PullRequestShortStatusMERGED,
		},
		statuses,
	)
}
