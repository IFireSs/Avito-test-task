package handlers

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/service"
	"context"
	"errors"
	"net/http"
	"strings"
)

type Server struct {
	prService   service.PRService
	teamService service.TeamService
	userService service.UserService

	adminToken string
}

var _ api.StrictServerInterface = (*Server)(nil)

func NewServer(
	prSvc service.PRService,
	teamSvc service.TeamService,
	userSvc service.UserService,
	adminToken string,
) *Server {
	return &Server{
		prService:   prSvc,
		teamService: teamSvc,
		userService: userSvc,
		adminToken:  adminToken,
	}
}

type adminTokenKey struct{}

func adminTokenFromContext(ctx context.Context) string {
	v := ctx.Value(adminTokenKey{})
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func AdminTokenMiddleware() api.StrictMiddlewareFunc {
	return func(next api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(
			ctx context.Context,
			w http.ResponseWriter,
			r *http.Request,
			request interface{},
		) (response interface{}, err error) {
			auth := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if strings.HasPrefix(auth, prefix) {
				token := strings.TrimSpace(auth[len(prefix):])
				if token != "" {
					ctx = context.WithValue(ctx, adminTokenKey{}, token)
				}
			}
			return next(ctx, w, r, request)
		}
	}
}

func makeError(code api.ErrorResponseErrorCode, msg string) api.ErrorResponse {
	return api.ErrorResponse{
		Error: struct {
			Code    api.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{
			Code:    code,
			Message: msg,
		},
	}
}

func mapDomainError(err error) (api.ErrorResponseErrorCode, int) {
	switch {
	case errors.Is(err, service.ErrTeamExists):
		return api.TEAMEXISTS, http.StatusBadRequest
	case errors.Is(err, service.ErrPRExists):
		return api.PREXISTS, http.StatusConflict
	case errors.Is(err, service.ErrPRMerged):
		return api.PRMERGED, http.StatusConflict
	case errors.Is(err, service.ErrReviewerNotAssigned):
		return api.NOTASSIGNED, http.StatusConflict
	case errors.Is(err, service.ErrNoCandidate):
		return api.NOCANDIDATE, http.StatusConflict
	case errors.Is(err, service.ErrNotFound):
		return api.NOTFOUND, http.StatusNotFound
	case errors.Is(err, service.ErrUnauthorized):
		return api.NOTFOUND, http.StatusUnauthorized
	default:
		return api.NOTFOUND, http.StatusInternalServerError
	}
}
