package service

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/repository"
	"context"
)

var (
	ErrTeamExists          = NewError("already exists")
	ErrPRExists            = NewError("PR id already exists")
	ErrPRMerged            = NewError("cannot reassign on merged PR")
	ErrReviewerNotAssigned = NewError("reviewer not assigned to this PR")
	ErrNoCandidate         = NewError("no replacement candidate")
	ErrNotFound            = NewError("resource not found")
	ErrUnauthorized        = NewError("unauthorized")
)

type DomainError struct {
	msg string
}

func (e DomainError) Error() string { return e.msg }

func NewError(msg string) error {
	return DomainError{msg: msg}
}

type TeamService interface {
	AddTeam(ctx context.Context, body api.PostTeamAddJSONRequestBody) (*api.Team, error)
	GetTeam(ctx context.Context, teamName string) (*api.Team, error)
}

type UserService interface {
	SetIsActive(ctx context.Context, body api.PostUsersSetIsActiveJSONRequestBody) (*api.User, error)
	GetReviews(ctx context.Context, userID string) ([]api.PullRequestShort, error)
	MassDeactivateTeamUsers(ctx context.Context, teamName string, userIDs []string) (*api.MassDeactivateResult, error)
}

type PRService interface {
	CreatePR(ctx context.Context, body api.PostPullRequestCreateJSONRequestBody) (*api.PullRequest, error)
	MergePR(ctx context.Context, body api.PostPullRequestMergeJSONRequestBody) (*api.PullRequest, error)
	ReassignReviewer(ctx context.Context, body api.PostPullRequestReassignJSONRequestBody) (*api.PullRequest, string, error)
	GetReviewerAssignments(ctx context.Context) ([]api.ReviewerStat, error)
}

func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository) TeamService {
	return &teamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func NewUserService(userRepo repository.UserRepository, prRepo repository.PRRepository) UserService {
	return &userService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func NewPRService(prRepo repository.PRRepository, userRepo repository.UserRepository) PRService {
	return &prService{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}
