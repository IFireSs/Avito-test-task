package http

import (
	"avito-autumn2025-internship/internal/api"
	"encoding/json"
	"fmt"
	"net/http"
)

func registerSwaggerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/swagger", swaggerUIHandler)
	mux.HandleFunc("/swagger/", swaggerUIHandler) // на всякий случай
	mux.HandleFunc("/swagger/openapi.json", swaggerJSONHandler)
}

func swaggerJSONHandler(w http.ResponseWriter, r *http.Request) {
	swagger, err := api.GetSwagger()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load swagger spec: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(swagger); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode swagger spec: %v", err), http.StatusInternalServerError)
		return
	}
}

func swaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	const page = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>PR Reviewer API – Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: '/swagger/openapi.json',
        dom_id: '#swagger-ui',
      });
    };
  </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, page)
}
