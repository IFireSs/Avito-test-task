package tests

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/service"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMassDeactivateTeamUsers_HappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	userRepo := newFakeUserRepo()
	prRepo := newFakePRRepo()

	userRepo.AddUser(api.User{
		UserId:   "u_lead",
		Username: "lead",
		TeamName: "backend",
		IsActive: true,
	})
	userRepo.AddUser(api.User{
		UserId:   "u_dev1",
		Username: "dev1",
		TeamName: "backend",
		IsActive: true,
	})
	userRepo.AddUser(api.User{
		UserId:   "u_dev2",
		Username: "dev2",
		TeamName: "backend",
		IsActive: true,
	})

	prRepo.AddPR(&api.PullRequest{
		PullRequestId:     "pr-1",
		PullRequestName:   "feature X",
		AuthorId:          "u_lead",
		Status:            api.PullRequestStatusOPEN,
		AssignedReviewers: []string{"u_dev1"},
	})
	prRepo.AddShortForReviewer("u_dev1", api.PullRequestShort{
		PullRequestId: "pr-1",
		Status:        api.PullRequestShortStatusOPEN,
	})

	userSvc := service.NewUserService(userRepo, prRepo)

	res, err := userSvc.MassDeactivateTeamUsers(ctx, "backend", []string{"u_dev1"})
	require.NoError(t, err)
	require.NotNil(t, res)

	require.Equal(t, 1, res.DeactivatedCount, "должны деактивировать одного пользователя")
	require.Equal(t, 1, res.ReassignedCount, "должен быть переназначен один PR")
	require.Equal(t, 0, res.NotReassignedCount, "все PR должны быть успешно переназначены")

	uDev1, err := userRepo.GetByID(ctx, "u_dev1")
	require.NoError(t, err)
	require.NotNil(t, uDev1)
	require.False(t, uDev1.IsActive, "u_dev1 должен быть деактивирован")

	require.Len(t, prRepo.replaceCalls, 1)
	call := prRepo.replaceCalls[0]
	require.Equal(t, "pr-1", call.PRID)
	require.Equal(t, "u_dev1", call.OldReviewerID)
	require.Equal(t, "u_dev2", call.NewReviewerID, "единственный допустимый кандидат — u_dev2")
}

func TestMassDeactivateTeamUsers_NoCandidates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	userRepo := newFakeUserRepo()
	prRepo := newFakePRRepo()

	userRepo.AddUser(api.User{
		UserId:   "u_author",
		Username: "author",
		TeamName: "data",
		IsActive: true,
	})
	userRepo.AddUser(api.User{
		UserId:   "u_rev",
		Username: "rev",
		TeamName: "data",
		IsActive: true,
	})

	prRepo.AddPR(&api.PullRequest{
		PullRequestId:     "pr-data-1",
		PullRequestName:   "data job",
		AuthorId:          "u_author",
		Status:            api.PullRequestStatusOPEN,
		AssignedReviewers: []string{"u_rev"},
	})
	prRepo.AddShortForReviewer("u_rev", api.PullRequestShort{
		PullRequestId: "pr-data-1",
		Status:        api.PullRequestShortStatusOPEN,
	})

	userSvc := service.NewUserService(userRepo, prRepo)

	res, err := userSvc.MassDeactivateTeamUsers(ctx, "data", []string{"u_rev"})
	require.NoError(t, err)
	require.NotNil(t, res)

	require.Equal(t, 1, res.DeactivatedCount, "ревьюер должен быть деактивирован")
	require.Equal(t, 0, res.ReassignedCount, "переназначений быть не должно")
	require.Equal(t, 1, res.NotReassignedCount, "один PR остался без безопасной замены")

	require.Len(t, prRepo.replaceCalls, 0)
}

func TestMassDeactivateTeamUsers_InvalidTeamName(t *testing.T) {
	t.Parallel()

	userRepo := newFakeUserRepo()
	prRepo := newFakePRRepo()

	userSvc := service.NewUserService(userRepo, prRepo)

	res, err := userSvc.MassDeactivateTeamUsers(context.Background(), "", []string{"u1"})
	require.Error(t, err)
	require.Nil(t, res)
	require.ErrorIs(t, err, service.ErrNotFound)
}
