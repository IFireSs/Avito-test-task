package handlers

import (
	"avito-autumn2025-internship/internal/api"
	"context"
	"net/http"
)

func (s *Server) PostPullRequestCreate(
	ctx context.Context,
	req api.PostPullRequestCreateRequestObject,
) (api.PostPullRequestCreateResponseObject, error) {
	if req.Body == nil {
		errResp := makeError(api.NOTFOUND, "request body is required")
		return api.PostPullRequestCreate404JSONResponse(errResp), nil
	}

	pr, err := s.prService.CreatePR(ctx, *req.Body)
	if err != nil {
		code, status := mapDomainError(err)
		errResp := makeError(code, err.Error())

		switch status {
		case http.StatusNotFound:
			return api.PostPullRequestCreate404JSONResponse(errResp), nil
		case http.StatusConflict:
			return api.PostPullRequestCreate409JSONResponse(errResp), nil
		default:
			return nil, err
		}
	}

	return api.PostPullRequestCreate201JSONResponse{
		Pr: pr,
	}, nil
}

func (s *Server) PostPullRequestMerge(
	ctx context.Context,
	req api.PostPullRequestMergeRequestObject,
) (api.PostPullRequestMergeResponseObject, error) {
	if req.Body == nil {
		errResp := makeError(api.NOTFOUND, "request body is required")
		return api.PostPullRequestMerge404JSONResponse(errResp), nil
	}

	pr, err := s.prService.MergePR(ctx, *req.Body)
	if err != nil {
		code, status := mapDomainError(err)
		errResp := makeError(code, err.Error())

		if status == http.StatusNotFound {
			return api.PostPullRequestMerge404JSONResponse(errResp), nil
		}
		return nil, err
	}

	return api.PostPullRequestMerge200JSONResponse{
		Pr: pr,
	}, nil
}

func (s *Server) PostPullRequestReassign(
	ctx context.Context,
	req api.PostPullRequestReassignRequestObject,
) (api.PostPullRequestReassignResponseObject, error) {
	if req.Body == nil {
		errResp := makeError(api.NOTFOUND, "request body is required")
		return api.PostPullRequestReassign404JSONResponse(errResp), nil
	}

	pr, replacedBy, err := s.prService.ReassignReviewer(ctx, *req.Body)
	if err != nil {
		code, status := mapDomainError(err)
		errResp := makeError(code, err.Error())

		switch status {
		case http.StatusNotFound:
			return api.PostPullRequestReassign404JSONResponse(errResp), nil
		case http.StatusConflict:
			return api.PostPullRequestReassign409JSONResponse(errResp), nil
		default:
			return nil, err
		}
	}

	return api.PostPullRequestReassign200JSONResponse{
		Pr:         *pr,
		ReplacedBy: replacedBy,
	}, nil
}
