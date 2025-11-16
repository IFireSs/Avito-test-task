package http

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/http/handlers"
	"avito-autumn2025-internship/internal/service"
	"log"
	nethttp "net/http"
	"time"
)

type loggingResponseWriter struct {
	nethttp.ResponseWriter
	statusCode int
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func loggingMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     nethttp.StatusOK,
		}

		start := time.Now()
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lrw.statusCode, duration)
	})
}

func NewRouter(
	prSvc service.PRService,
	teamSvc service.TeamService,
	userSvc service.UserService,
	adminToken string,
) nethttp.Handler {
	srv := handlers.NewServer(prSvc, teamSvc, userSvc, adminToken)

	strict := api.NewStrictHandler(srv, []api.StrictMiddlewareFunc{
		handlers.AdminTokenMiddleware(),
	})

	mux := nethttp.NewServeMux()
	registerSwaggerRoutes(mux)

	handler := api.HandlerFromMux(strict, mux)
	return loggingMiddleware(handler)
}
