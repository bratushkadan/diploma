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
