package service

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/repository"
	"context"
)

type userService struct {
	userRepo repository.UserRepository
	prRepo   repository.PRRepository
}

func (s *userService) SetIsActive(ctx context.Context, body api.PostUsersSetIsActiveJSONRequestBody) (*api.User, error) {
	if body.UserId == "" {
		return nil, ErrNotFound
	}

	user, err := s.userRepo.GetByID(ctx, body.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}

	user, err = s.userRepo.SetIsActive(ctx, body.UserId, body.IsActive)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetReviews(ctx context.Context, userID string) ([]api.PullRequestShort, error) {
	if userID == "" {
		return nil, ErrNotFound
	}

	prs, err := s.prRepo.ListShortByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (s *userService) MassDeactivateTeamUsers(
	ctx context.Context,
	teamName string,
	userIDs []string,
) (*api.MassDeactivateResult, error) {
	if teamName == "" {
		return nil, ErrNotFound
	}
	if len(userIDs) == 0 {
		return &api.MassDeactivateResult{}, nil
	}

	teamMembers, err := s.userRepo.ListByTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}
	if len(teamMembers) == 0 {
		return nil, ErrNotFound
	}

	membersByID := make(map[string]api.User, len(teamMembers))
	for _, u := range teamMembers {
		membersByID[u.UserId] = u
	}

	targetSet := make(map[string]struct{}, len(userIDs))
	var targets []string
	for _, id := range userIDs {
		if _, ok := membersByID[id]; !ok {
			continue
		}
		if _, seen := targetSet[id]; seen {
			continue
		}
		targetSet[id] = struct{}{}
		targets = append(targets, id)
	}

	if len(targets) == 0 {
		return &api.MassDeactivateResult{}, nil
	}

	res := &api.MassDeactivateResult{}

	for _, id := range targets {
		u := membersByID[id]
		if !u.IsActive {
			continue
		}

		updated, err := s.userRepo.SetIsActive(ctx, id, false)
		if err != nil {
			return nil, err
		}
		if updated == nil {
			continue
		}

		res.DeactivatedCount++

		u.IsActive = false
		membersByID[id] = *updated
	}

	activeCandidates := make([]api.User, 0, len(membersByID))
	for _, u := range membersByID {
		if u.IsActive {
			activeCandidates = append(activeCandidates, u)
		}
	}

	if len(activeCandidates) == 0 {
		return res, nil
	}
	for _, deactivatedID := range targets {
		prs, err := s.prRepo.ListShortByReviewer(ctx, deactivatedID)
		if err != nil {
			return nil, err
		}

		for _, short := range prs {
			if short.Status != api.PullRequestShortStatusOPEN {
				continue
			}

			pr, err := s.prRepo.GetByID(ctx, short.PullRequestId)
			if err != nil {
				return nil, err
			}
			if pr == nil {
				continue
			}

			found := false
			for _, r := range pr.AssignedReviewers {
				if r == deactivatedID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			exclude := make(map[string]struct{}, len(pr.AssignedReviewers)+1)
			exclude[pr.AuthorId] = struct{}{}
			for _, r := range pr.AssignedReviewers {
				exclude[r] = struct{}{}
			}

			localCandidates := make([]api.User, 0, len(activeCandidates))
			for _, u := range activeCandidates {
				if _, skip := exclude[u.UserId]; skip {
					continue
				}
				localCandidates = append(localCandidates, u)
			}

			if len(localCandidates) == 0 {
				res.NotReassignedCount++
				continue
			}

			newIDs := chooseRandomReviewers(localCandidates, 1)
			if len(newIDs) == 0 {
				res.NotReassignedCount++
				continue
			}
			newID := newIDs[0]

			if err := s.prRepo.ReplaceReviewer(ctx, pr.PullRequestId, deactivatedID, newID); err != nil {
				return nil, err
			}

			res.ReassignedCount++
		}
	}

	return res, nil
}
