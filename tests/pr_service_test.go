package tests

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/service"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPRService_CreatePR_Success_AssignsTwoReviewers(t *testing.T) {
	ctx := context.Background()

	userRepo := newFakeUserRepo()
	prRepo := newFakePRRepo()

	userRepo.AddUser(api.User{
		UserId:   "u_author",
		Username: "author",
		TeamName: "backend",
		IsActive: true,
	})
	userRepo.AddUser(api.User{
		UserId:   "u_r1",
		Username: "rev1",
		TeamName: "backend",
		IsActive: true,
	})
	userRepo.AddUser(api.User{
		UserId:   "u_r2",
		Username: "rev2",
		TeamName: "backend",
		IsActive: true,
	})

	prSvc := service.NewPRService(prRepo, userRepo)

	body := api.PostPullRequestCreateJSONRequestBody{
		PullRequestId:   "pr-1",
		PullRequestName: "Implement feature X",
		AuthorId:        "u_author",
	}

	pr, err := prSvc.CreatePR(ctx, body)
	require.NoError(t, err)
	require.NotNil(t, pr)

	require.Equal(t, "pr-1", pr.PullRequestId)
	require.Equal(t, "u_author", pr.AuthorId)
	require.Equal(t, api.PullRequestStatusOPEN, pr.Status)
	require.NotNil(t, pr.CreatedAt)
	require.Nil(t, pr.MergedAt)

	require.Len(t, pr.AssignedReviewers, 2)
	require.ElementsMatch(t, []string{"u_r1", "u_r2"}, pr.AssignedReviewers)

	stored, err := prRepo.GetByID(ctx, "pr-1")
	require.NoError(t, err)
	require.NotNil(t, stored)
	require.Equal(t, pr.PullRequestId, stored.PullRequestId)
}

func TestPRService_CreatePR_PRAlreadyExists(t *testing.T) {
	ctx := context.Background()

	userRepo := newFakeUserRepo()
	prRepo := newFakePRRepo()

	prRepo.AddPR(&api.PullRequest{
		PullRequestId: "pr-1",
	})

	prSvc := service.NewPRService(prRepo, userRepo)

	body := api.PostPullRequestCreateJSONRequestBody{
		PullRequestId:   "pr-1",
		PullRequestName: "Duplicate",
		AuthorId:        "u_author",
	}

	pr, err := prSvc.CreatePR(ctx, body)
	require.Error(t, err)
	require.Nil(t, pr)
	require.Equal(t, service.ErrPRExists, err)
}

func TestPRService_ReassignReviewer_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := newFakeUserRepo()
	prRepo := newFakePRRepo()

	userRepo.AddUser(api.User{
		UserId:   "u_author",
		Username: "author",
		TeamName: "backend",
		IsActive: true,
	})
	userRepo.AddUser(api.User{
		UserId:   "u_old",
		Username: "old_reviewer",
		TeamName: "backend",
		IsActive: true,
	})
	userRepo.AddUser(api.User{
		UserId:   "u_new",
		Username: "new_reviewer",
		TeamName: "backend",
		IsActive: true,
	})

	prRepo.AddPR(&api.PullRequest{
		PullRequestId:     "pr-1",
		PullRequestName:   "Some PR",
		AuthorId:          "u_author",
		Status:            api.PullRequestStatusOPEN,
		AssignedReviewers: []string{"u_old"},
	})

	prSvc := service.NewPRService(prRepo, userRepo)

	body := api.PostPullRequestReassignJSONRequestBody{
		PullRequestId: "pr-1",
		OldUserId:     "u_old",
	}

	pr, replacedBy, err := prSvc.ReassignReviewer(ctx, body)
	require.NoError(t, err)
	require.NotNil(t, pr)

	require.Equal(t, "u_new", replacedBy)
	require.Len(t, pr.AssignedReviewers, 1)
	require.Equal(t, "u_new", pr.AssignedReviewers[0])

	require.Len(t, prRepo.replaceCalls, 1)
	call := prRepo.replaceCalls[0]
	require.Equal(t, "pr-1", call.PRID)
	require.Equal(t, "u_old", call.OldReviewerID)
	require.Equal(t, "u_new", call.NewReviewerID)
}

func TestPRService_ReassignReviewer_PRIsMerged(t *testing.T) {
	ctx := context.Background()

	userRepo := newFakeUserRepo()
	prRepo := newFakePRRepo()

	prRepo.AddPR(&api.PullRequest{
		PullRequestId:     "pr-1",
		PullRequestName:   "Merged PR",
		AuthorId:          "u_author",
		Status:            api.PullRequestStatusMERGED,
		AssignedReviewers: []string{"u_old"},
	})

	prSvc := service.NewPRService(prRepo, userRepo)

	body := api.PostPullRequestReassignJSONRequestBody{
		PullRequestId: "pr-1",
		OldUserId:     "u_old",
	}

	pr, replacedBy, err := prSvc.ReassignReviewer(ctx, body)
	require.Error(t, err)
	require.Nil(t, pr)
	require.Equal(t, "", replacedBy)
	require.Equal(t, service.ErrPRMerged, err)

	require.Len(t, prRepo.replaceCalls, 0)
}
