package handlers

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/service"
	"context"
	"net/http"
)

func (s *Server) PostTeamAdd(
	ctx context.Context,
	req api.PostTeamAddRequestObject,
) (api.PostTeamAddResponseObject, error) {
	if req.Body == nil {
		errResp := makeError(api.TEAMEXISTS, "request body is required")
		return api.PostTeamAdd400JSONResponse(errResp), nil
	}

	team, err := s.teamService.AddTeam(ctx, *req.Body)
	if err != nil {
		code, status := mapDomainError(err)
		errResp := makeError(code, req.Body.TeamName+" "+err.Error())

		switch status {
		case http.StatusBadRequest:
			return api.PostTeamAdd400JSONResponse(errResp), nil
		default:
			return nil, err
		}
	}

	return api.PostTeamAdd201JSONResponse{
		Team: team,
	}, nil
}

func (s *Server) GetTeamGet(
	ctx context.Context,
	req api.GetTeamGetRequestObject,
) (api.GetTeamGetResponseObject, error) {
	teamName := string(req.Params.TeamName)

	team, err := s.teamService.GetTeam(ctx, teamName)
	if err != nil {
		code, status := mapDomainError(err)
		errResp := makeError(code, err.Error())

		if status == http.StatusNotFound {
			return api.GetTeamGet404JSONResponse(errResp), nil
		}
		return nil, err
	}
	return api.GetTeamGet200JSONResponse(*team), nil
}

func (s *Server) PostTeamMassDeactivate(
	ctx context.Context,
	req api.PostTeamMassDeactivateRequestObject,
) (api.PostTeamMassDeactivateResponseObject, error) {
	if req.Body == nil {
		errResp := makeError(api.NOTFOUND, "request body is required")
		return api.PostTeamMassDeactivate404JSONResponse(errResp), nil
	}

	body := req.Body

	if s.adminToken != "" {
		token := adminTokenFromContext(ctx)
		if token == "" || token != s.adminToken {
			err := service.ErrUnauthorized
			code, _ := mapDomainError(err)
			errResp := makeError(code, err.Error())
			return api.PostTeamMassDeactivate401JSONResponse(errResp), nil
		}
	}

	result, err := s.userService.MassDeactivateTeamUsers(ctx, body.TeamName, body.UserIds)
	if err != nil {
		code, status := mapDomainError(err)
		errResp := makeError(code, err.Error())

		switch status {
		case http.StatusNotFound:
			return api.PostTeamMassDeactivate404JSONResponse(errResp), nil
		case http.StatusUnauthorized:
			return api.PostTeamMassDeactivate401JSONResponse(errResp), nil
		default:
			return nil, err
		}
	}

	resp := &api.MassDeactivateResult{
		DeactivatedCount:   result.DeactivatedCount,
		ReassignedCount:    result.ReassignedCount,
		NotReassignedCount: result.NotReassignedCount,
	}

	return api.PostTeamMassDeactivate200JSONResponse{
		Result: resp,
	}, nil
}
