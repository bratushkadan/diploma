package frontend

import (
	"net/http"

	"github.com/bratushkadan/floral/internal/auth/domain"
	"github.com/go-chi/chi/v5"
)

type HttpImpl struct {
	auth *domain.AuthService
}

func (f *HttpImpl) Start(auth *domain.AuthService) error {
	f.auth = auth

	r := chi.NewRouter()

	apiRouter := chi.NewRouter()

	v1Router := chi.NewRouter()

	apiRouter.Mount("/v1", v1Router)

	r.Mount("/api", apiRouter)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	return http.ListenAndServe(":8080", r)
}
