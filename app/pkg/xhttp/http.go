package xhttp

import (
	"context"
	"net/http"
)

func HandleReadiness(ctx context.Context) http.HandlerFunc {
	f := func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-ctx.Done():
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"message": "shutting down"}`))
		default:
			w.Write([]byte(`{"message": "ok"}`))
		}
	}
	return http.HandlerFunc(f)
}
func HandleNotFound() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"code": 0, "message": "handler not found"}]}`))
	})
}

func NewErrorResponse(errs ...ErrorResponseErr) ErrorResponse {
	return ErrorResponse{
		Errors: errs,
	}
}

type ErrorResponse struct {
	Errors []ErrorResponseErr `json:"errors"`
}

type ErrorResponseErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
