package tests

import (
	"avito-autumn2025-internship/internal/api"
	nethttp "avito-autumn2025-internship/internal/http"
	"avito-autumn2025-internship/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTP_CreatePullRequest_HappyPath(t *testing.T) {
	t.Parallel()

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
	userRepo.AddUser(api.User{
		UserId:   "u_r3",
		Username: "rev3",
		TeamName: "backend",
		IsActive: true,
	})

	prSvc := service.NewPRService(prRepo, userRepo)
	teamSvc := newTeamServiceStub()
	userSvc := service.NewUserService(userRepo, prRepo)

	const adminToken = ""

	handler := nethttp.NewRouter(prSvc, teamSvc, userSvc, adminToken)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	body := api.PostPullRequestCreateJSONRequestBody{
		PullRequestId:   "pr-100",
		PullRequestName: "Implement feature Y",
		AuthorId:        "u_author",
	}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(
		http.MethodPost,
		ts.URL+"/pullRequest/create",
		bytes.NewReader(bodyBytes),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody struct {
		PR *api.PullRequest `json:"pr"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.NotNil(t, respBody.PR)

	pr := respBody.PR

	require.Equal(t, "pr-100", pr.PullRequestId)
	require.Equal(t, "u_author", pr.AuthorId)
	require.Equal(t, api.PullRequestStatusOPEN, pr.Status)
	require.NotNil(t, pr.CreatedAt)
	require.Nil(t, pr.MergedAt)

	require.Len(t, pr.AssignedReviewers, 2)

	for _, rID := range pr.AssignedReviewers {
		require.NotEqual(t, "u_author", rID)
	}

	valid := map[string]bool{
		"u_r1": true,
		"u_r2": true,
		"u_r3": true,
	}
	for _, rID := range pr.AssignedReviewers {
		require.True(t, valid[rID], "unexpected reviewer id %s", rID)
	}
	stored, err := prRepo.GetByID(ctx, "pr-100")
	require.NoError(t, err)
	require.NotNil(t, stored)
	require.Equal(t, pr.PullRequestId, stored.PullRequestId)
}
