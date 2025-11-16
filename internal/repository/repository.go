package repository

import (
	"avito-autumn2025-internship/internal/api"
	"context"
	"time"
)

type ReviewerAssignmentsStat struct {
	UserID string
	Count  int64
}

type TeamRepository interface {
	Create(ctx context.Context, teamName string) error
	Exists(ctx context.Context, teamName string) (bool, error)
}

type UserRepository interface {
	UpsertTeamMembers(ctx context.Context, teamName string, members []api.TeamMember) ([]api.User, error)

	ListByTeam(ctx context.Context, teamName string) ([]api.User, error)
	GetByID(ctx context.Context, userID string) (*api.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error)
	ListActiveByTeam(ctx context.Context, teamName string) ([]api.User, error)
}

type PRRepository interface {
	Create(ctx context.Context, pr *api.PullRequest) error
	GetByID(ctx context.Context, prID string) (*api.PullRequest, error)

	SetMerged(ctx context.Context, prID string, mergedAt time.Time) (*api.PullRequest, error)
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error

	ListReviewers(ctx context.Context, prID string) ([]string, error)
	SetReviewers(ctx context.Context, prID string, reviewers []string) error

	ListShortByReviewer(ctx context.Context, reviewerID string) ([]api.PullRequestShort, error)
	GetReviewerAssignmentsStats(ctx context.Context) ([]ReviewerAssignmentsStat, error)
}
