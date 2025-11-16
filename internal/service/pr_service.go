package service

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/repository"
	"context"
	"math/rand"
	"time"
)

type prService struct {
	prRepo   repository.PRRepository
	userRepo repository.UserRepository
}

func (s *prService) CreatePR(ctx context.Context, body api.PostPullRequestCreateJSONRequestBody) (*api.PullRequest, error) {
	if body.PullRequestId == "" || body.PullRequestName == "" || body.AuthorId == "" {
		return nil, ErrNotFound
	}

	existing, err := s.prRepo.GetByID(ctx, body.PullRequestId)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPRExists
	}

	author, err := s.userRepo.GetByID(ctx, body.AuthorId)
	if err != nil {
		return nil, err
	}
	if author == nil {
		return nil, ErrNotFound
	}

	teamName := author.TeamName
	if teamName == "" {
		return nil, ErrNotFound
	}

	teamMembers, err := s.userRepo.ListActiveByTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}

	candidates := make([]api.User, 0, len(teamMembers))
	for _, u := range teamMembers {
		if u.UserId == author.UserId {
			continue
		}
		candidates = append(candidates, u)
	}

	assigned := chooseRandomReviewers(candidates, 2)

	now := time.Now().UTC()
	var mergedAt *time.Time

	pr := &api.PullRequest{
		PullRequestId:     body.PullRequestId,
		PullRequestName:   body.PullRequestName,
		AuthorId:          body.AuthorId,
		Status:            api.PullRequestStatusOPEN,
		CreatedAt:         &now,
		MergedAt:          mergedAt,
		AssignedReviewers: assigned,
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func chooseRandomReviewers(users []api.User, max int) []string {
	if len(users) == 0 || max <= 0 {
		return nil
	}

	idxs := rand.Perm(len(users))
	if len(idxs) > max {
		idxs = idxs[:max]
	}

	result := make([]string, 0, len(idxs))
	for _, i := range idxs {
		result = append(result, users[i].UserId)
	}
	return result
}

func (s *prService) MergePR(ctx context.Context, body api.PostPullRequestMergeJSONRequestBody) (*api.PullRequest, error) {
	if body.PullRequestId == "" {
		return nil, ErrNotFound
	}

	pr, err := s.prRepo.GetByID(ctx, body.PullRequestId)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, ErrNotFound
	}

	if pr.Status == api.PullRequestStatusMERGED {
		return pr, nil
	}

	now := time.Now().UTC()
	updated, err := s.prRepo.SetMerged(ctx, body.PullRequestId, now)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *prService) ReassignReviewer(
	ctx context.Context,
	body api.PostPullRequestReassignJSONRequestBody,
) (*api.PullRequest, string, error) {
	if body.PullRequestId == "" || body.OldUserId == "" {
		return nil, "", ErrNotFound
	}

	pr, err := s.prRepo.GetByID(ctx, body.PullRequestId)
	if err != nil {
		return nil, "", err
	}
	if pr == nil {
		return nil, "", ErrNotFound
	}

	if pr.Status == api.PullRequestStatusMERGED {
		return nil, "", ErrPRMerged
	}

	reviewers := pr.AssignedReviewers
	found := false
	for _, r := range reviewers {
		if r == body.OldUserId {
			found = true
			break
		}
	}
	if !found {
		return nil, "", ErrReviewerNotAssigned
	}

	oldUser, err := s.userRepo.GetByID(ctx, body.OldUserId)
	if err != nil {
		return nil, "", err
	}
	if oldUser == nil {
		return nil, "", ErrNotFound
	}
	teamName := oldUser.TeamName
	if teamName == "" {
		return nil, "", ErrNotFound
	}

	teamMembers, err := s.userRepo.ListActiveByTeam(ctx, teamName)
	if err != nil {
		return nil, "", err
	}

	exclude := map[string]struct{}{
		pr.AuthorId: {},
	}
	for _, r := range reviewers {
		exclude[r] = struct{}{}
	}

	candidates := make([]api.User, 0, len(teamMembers))
	for _, u := range teamMembers {
		if _, skip := exclude[u.UserId]; skip {
			continue
		}
		candidates = append(candidates, u)
	}

	if len(candidates) == 0 {
		return nil, "", ErrNoCandidate
	}

	newID := chooseRandomReviewers(candidates, 1)[0]

	if err := s.prRepo.ReplaceReviewer(ctx, pr.PullRequestId, body.OldUserId, newID); err != nil {
		return nil, "", err
	}

	for i, r := range pr.AssignedReviewers {
		if r == body.OldUserId {
			pr.AssignedReviewers[i] = newID
			break
		}
	}

	return pr, newID, nil
}

func (s *prService) GetReviewerAssignments(ctx context.Context) ([]api.ReviewerStat, error) {
	stats, err := s.prRepo.GetReviewerAssignmentsStats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]api.ReviewerStat, 0, len(stats))
	for _, st := range stats {
		result = append(result, api.ReviewerStat{
			UserId:        st.UserID,
			AssignedCount: st.Count,
		})
	}
	return result, nil
}
