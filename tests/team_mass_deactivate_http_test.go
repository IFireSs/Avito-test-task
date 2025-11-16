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

func TestHTTP_TeamMassDeactivate_HappyPath(t *testing.T) {
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

	prSvc := newPRServiceStub()
	teamSvc := newTeamServiceStub()

	const adminToken = "secret-admin"

	handler := nethttp.NewRouter(prSvc, teamSvc, userSvc, adminToken)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	reqBody := map[string]interface{}{
		"team_name": "backend",
		"user_ids":  []string{"u_dev1"},
	}

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(
		http.MethodPost,
		ts.URL+"/team/massDeactivate",
		bytes.NewReader(bodyBytes),
	)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody struct {
		Result *api.MassDeactivateResult `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.NotNil(t, respBody.Result)

	require.Equal(t, 1, respBody.Result.DeactivatedCount)
	require.Equal(t, 1, respBody.Result.ReassignedCount)
	require.Equal(t, 0, respBody.Result.NotReassignedCount)

	uDev1, err := userRepo.GetByID(ctx, "u_dev1")
	require.NoError(t, err)
	require.NotNil(t, uDev1)
	require.False(t, uDev1.IsActive, "u_dev1 должен быть деактивирован")

	require.Len(t, prRepo.replaceCalls, 1)
	call := prRepo.replaceCalls[0]
	require.Equal(t, "pr-1", call.PRID)
	require.Equal(t, "u_dev1", call.OldReviewerID)
	require.Equal(t, "u_dev2", call.NewReviewerID)
}

func TestHTTP_TeamMassDeactivate_UnauthorizedWithoutToken(t *testing.T) {
	t.Parallel()

	userRepo := newFakeUserRepo()
	prRepo := newFakePRRepo()

	userSvc := service.NewUserService(userRepo, prRepo)
	prSvc := newPRServiceStub()
	teamSvc := newTeamServiceStub()

	const adminToken = "secret-admin"

	handler := nethttp.NewRouter(prSvc, teamSvc, userSvc, adminToken)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	reqBody := map[string]interface{}{
		"team_name": "backend",
		"user_ids":  []string{"u_dev1"},
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(
		http.MethodPost,
		ts.URL+"/team/massDeactivate",
		bytes.NewReader(bodyBytes),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
