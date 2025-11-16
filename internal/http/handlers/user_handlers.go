package handlers

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/service"
	"context"
	"errors"
	"net/http"
)

func (s *Server) PostUsersSetIsActive(
	ctx context.Context,
	req api.PostUsersSetIsActiveRequestObject,
) (api.PostUsersSetIsActiveResponseObject, error) {
	if s.adminToken != "" {
		token := adminTokenFromContext(ctx)
		if token == "" || token != s.adminToken {
			err := service.ErrUnauthorized
			code, _ := mapDomainError(err)
			errResp := makeError(code, err.Error())
			return api.PostUsersSetIsActive401JSONResponse(errResp), nil
		}
	}

	if req.Body == nil {
		errResp := makeError(api.NOTFOUND, "request body is required")
		return api.PostUsersSetIsActive404JSONResponse(errResp), nil
	}

	user, err := s.userService.SetIsActive(ctx, *req.Body)
	if err != nil {
		code, status := mapDomainError(err)
		errResp := makeError(code, err.Error())

		switch status {
		case http.StatusNotFound:
			return api.PostUsersSetIsActive404JSONResponse(errResp), nil
		case http.StatusUnauthorized:
			return api.PostUsersSetIsActive401JSONResponse(errResp), nil
		default:
			return nil, err
		}
	}

	return api.PostUsersSetIsActive200JSONResponse{
		User: user,
	}, nil
}

func (s *Server) GetUsersGetReview(
	ctx context.Context,
	req api.GetUsersGetReviewRequestObject,
) (api.GetUsersGetReviewResponseObject, error) {
	userID := string(req.Params.UserId)

	prs, err := s.userService.GetReviews(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return api.GetUsersGetReview200JSONResponse{
				UserId:       userID,
				PullRequests: []api.PullRequestShort{},
			}, nil
		}
		return nil, err
	}

	return api.GetUsersGetReview200JSONResponse{
		UserId:       userID,
		PullRequests: prs,
	}, nil
}
