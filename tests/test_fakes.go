package tests

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/repository"
	"avito-autumn2025-internship/internal/service"
	"context"
	"time"
)

type fakeUserRepo struct {
	users map[string]*api.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users: make(map[string]*api.User),
	}
}

func (r *fakeUserRepo) AddUser(u api.User) {
	uCopy := u
	r.users[u.UserId] = &uCopy
}

func (r *fakeUserRepo) UpsertTeamMembers(
	_ context.Context,
	teamName string,
	members []api.TeamMember,
) ([]api.User, error) {
	var res []api.User
	for _, m := range members {
		u := api.User{
			UserId:   m.UserId,
			Username: m.Username,
			TeamName: teamName,
			IsActive: true,
		}
		r.AddUser(u)
		res = append(res, u)
	}
	return res, nil
}

func (r *fakeUserRepo) ListByTeam(_ context.Context, teamName string) ([]api.User, error) {
	var res []api.User
	for _, u := range r.users {
		if u.TeamName == teamName {
			res = append(res, *u)
		}
	}
	return res, nil
}

func (r *fakeUserRepo) GetByID(_ context.Context, userID string) (*api.User, error) {
	u, ok := r.users[userID]
	if !ok {
		return nil, nil
	}
	uCopy := *u
	return &uCopy, nil
}

func (r *fakeUserRepo) SetIsActive(_ context.Context, userID string, isActive bool) (*api.User, error) {
	u, ok := r.users[userID]
	if !ok {
		return nil, nil
	}
	u.IsActive = isActive
	uCopy := *u
	return &uCopy, nil
}

func (r *fakeUserRepo) ListActiveByTeam(_ context.Context, teamName string) ([]api.User, error) {
	var res []api.User
	for _, u := range r.users {
		if u.TeamName == teamName && u.IsActive {
			res = append(res, *u)
		}
	}
	return res, nil
}

var _ repository.UserRepository = (*fakeUserRepo)(nil)

type fakePRRepo struct {
	prs             map[string]*api.PullRequest
	shortByReviewer map[string][]api.PullRequestShort

	replaceCalls []struct {
		PRID          string
		OldReviewerID string
		NewReviewerID string
	}
}

func newFakePRRepo() *fakePRRepo {
	return &fakePRRepo{
		prs:             make(map[string]*api.PullRequest),
		shortByReviewer: make(map[string][]api.PullRequestShort),
	}
}

func (r *fakePRRepo) AddPR(pr *api.PullRequest) {
	cp := *pr
	r.prs[pr.PullRequestId] = &cp
}

func (r *fakePRRepo) AddShortForReviewer(reviewerID string, prs ...api.PullRequestShort) {
	r.shortByReviewer[reviewerID] = append(r.shortByReviewer[reviewerID], prs...)
}

func (r *fakePRRepo) Create(_ context.Context, pr *api.PullRequest) error {
	if _, exists := r.prs[pr.PullRequestId]; exists {
	}
	r.AddPR(pr)
	return nil
}

func (r *fakePRRepo) GetByID(_ context.Context, prID string) (*api.PullRequest, error) {
	pr, ok := r.prs[prID]
	if !ok {
		return nil, nil
	}
	cp := *pr
	return &cp, nil
}

func (r *fakePRRepo) SetMerged(_ context.Context, prID string, mergedAt time.Time) (*api.PullRequest, error) {
	pr, ok := r.prs[prID]
	if !ok {
		return nil, nil
	}
	cp := *pr
	cp.Status = api.PullRequestStatusMERGED
	cp.MergedAt = &mergedAt
	r.prs[prID] = &cp
	return &cp, nil
}

func (r *fakePRRepo) ReplaceReviewer(_ context.Context, prID, oldReviewerID, newReviewerID string) error {
	r.replaceCalls = append(r.replaceCalls, struct {
		PRID          string
		OldReviewerID string
		NewReviewerID string
	}{
		PRID:          prID,
		OldReviewerID: oldReviewerID,
		NewReviewerID: newReviewerID,
	})

	if pr, ok := r.prs[prID]; ok {
		for i, rid := range pr.AssignedReviewers {
			if rid == oldReviewerID {
				pr.AssignedReviewers[i] = newReviewerID
			}
		}
	}

	return nil
}

func (r *fakePRRepo) ListReviewers(_ context.Context, prID string) ([]string, error) {
	pr, ok := r.prs[prID]
	if !ok {
		return nil, nil
	}
	out := make([]string, len(pr.AssignedReviewers))
	copy(out, pr.AssignedReviewers)
	return out, nil
}

func (r *fakePRRepo) SetReviewers(_ context.Context, prID string, reviewers []string) error {
	pr, ok := r.prs[prID]
	if !ok {
		return nil
	}
	cp := make([]string, len(reviewers))
	copy(cp, reviewers)
	pr.AssignedReviewers = cp
	return nil
}

func (r *fakePRRepo) ListShortByReviewer(_ context.Context, reviewerID string) ([]api.PullRequestShort, error) {
	out := r.shortByReviewer[reviewerID]
	cp := make([]api.PullRequestShort, len(out))
	copy(cp, out)
	return cp, nil
}

func (r *fakePRRepo) GetReviewerAssignmentsStats(_ context.Context) ([]repository.ReviewerAssignmentsStat, error) {
	return nil, nil
}

var _ repository.PRRepository = (*fakePRRepo)(nil)

type prServiceStub struct{}
type teamServiceStub struct{}

func newPRServiceStub() service.PRService     { return &prServiceStub{} }
func newTeamServiceStub() service.TeamService { return &teamServiceStub{} }

func (*prServiceStub) CreatePR(ctx context.Context, body api.PostPullRequestCreateJSONRequestBody) (*api.PullRequest, error) {
	panic("not implemented")
}

func (*prServiceStub) MergePR(ctx context.Context, body api.PostPullRequestMergeJSONRequestBody) (*api.PullRequest, error) {
	panic("not implemented")
}

func (*prServiceStub) ReassignReviewer(ctx context.Context, body api.PostPullRequestReassignJSONRequestBody) (*api.PullRequest, string, error) {
	panic("not implemented")
}

func (*prServiceStub) GetReviewerAssignments(ctx context.Context) ([]api.ReviewerStat, error) {
	panic("not implemented")
}

func (*teamServiceStub) AddTeam(ctx context.Context, body api.PostTeamAddJSONRequestBody) (*api.Team, error) {
	panic("not implemented")
}

func (*teamServiceStub) GetTeam(ctx context.Context, teamName string) (*api.Team, error) {
	panic("not implemented")
}

var _ service.PRService = (*prServiceStub)(nil)
var _ service.TeamService = (*teamServiceStub)(nil)
