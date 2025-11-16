package handlers

import (
	"avito-autumn2025-internship/internal/api"
	"context"
)

func (s *Server) GetStatsReviewerAssignments(
	ctx context.Context,
	_ api.GetStatsReviewerAssignmentsRequestObject,
) (api.GetStatsReviewerAssignmentsResponseObject, error) {
	stats, err := s.prService.GetReviewerAssignments(ctx)
	if err != nil {
		return nil, err
	}

	return api.GetStatsReviewerAssignments200JSONResponse{
		Stats: stats,
	}, nil
}
