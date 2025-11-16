package service

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/repository"
	"context"
)

type teamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

func (s *teamService) AddTeam(ctx context.Context, body api.PostTeamAddJSONRequestBody) (*api.Team, error) {
	if body.TeamName == "" {
		return nil, ErrNotFound
	}

	exists, err := s.teamRepo.Exists(ctx, body.TeamName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTeamExists
	}

	if err := s.teamRepo.Create(ctx, body.TeamName); err != nil {
		return nil, ErrTeamExists
	}

	users, err := s.userRepo.UpsertTeamMembers(ctx, body.TeamName, body.Members)
	if err != nil {
		return nil, err
	}

	members := make([]api.TeamMember, 0, len(users))
	for _, u := range users {
		members = append(members, api.TeamMember{
			UserId:   u.UserId,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	team := &api.Team{
		TeamName: body.TeamName,
		Members:  members,
	}

	return team, nil
}

func (s *teamService) GetTeam(ctx context.Context, teamName string) (*api.Team, error) {
	if teamName == "" {
		return nil, ErrNotFound
	}

	exists, err := s.teamRepo.Exists(ctx, teamName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	users, err := s.userRepo.ListByTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}

	members := make([]api.TeamMember, 0, len(users))
	for _, u := range users {
		members = append(members, api.TeamMember{
			UserId:   u.UserId,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	team := &api.Team{
		TeamName: teamName,
		Members:  members,
	}
	return team, nil
}
