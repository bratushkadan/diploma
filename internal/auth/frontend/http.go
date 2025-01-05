package frontend

import (
	"encoding/json"
	"errors"
	"log"
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

	usersRouter := chi.NewRouter()

	v1Router.Mount("/users", usersRouter)

	usersRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("i am a teapot"))
	})
	usersRouter.Post("/:register", http.HandlerFunc(f.RegisterUserHandler))
	usersRouter.Post("/:authenticate", http.HandlerFunc(f.AuthenticateHandler))
	usersRouter.Post("/:renewRefreshToken", http.HandlerFunc(f.RenewRefreshTokenHandler))

	r.Mount("/api", apiRouter)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	return http.ListenAndServe(":8080", r)
}

func (f *HttpImpl) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors":[{"code": 123, "message": "bad request body"}]}`))
		return
	}

	user, err := f.auth.CreateCustomer(r.Context(), domain.CreateCustomerReq{
		CreateUserReq: domain.CreateUserReq{
			Name:     reqData.Name,
			Password: reqData.Password,
			Email:    reqData.Email,
		},
	})
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"code": 1, "message": "internal server error"}]}`))
		return
	}

	if err := json.NewEncoder(w).Encode(&struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}{
		Id:   user.Id,
		Name: user.Name,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"code": 1, "message": "internal server error"}]}`))
		return
	}
}

func (f *HttpImpl) AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors":[{"code": 123, "message": "bad request body"}]}`))
		return
	}

	refreshToken, err := f.auth.Authenticate(r.Context(), reqData.Email, reqData.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"errors":[{"code": 2, "message": "bad user credentials"}]}`))
			return
		}
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"code": 1, "message": "internal server error"}]}`))
		return
	}

	if err := json.NewEncoder(w).Encode(&struct {
		RefreshToken string `json:"refresh_token"`
	}{
		RefreshToken: refreshToken,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"code": 1, "message": "internal server error"}]}`))
		return
	}
}

func (f *HttpImpl) RenewRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil || reqData.RefreshToken == "" {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors":[{"code": 123, "message": "bad request body"}]}`))
		return
	}

	refreshToken, err := f.auth.RenewRefreshToken(r.Context(), reqData.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"errors":[{"code": 3, "message": "invalid refresh token"}]}`))
			return
		}
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"code": 1, "message": "internal server error"}]}`))
		return
	}

	if err := json.NewEncoder(w).Encode(&struct {
		RefreshToken string `json:"refresh_token"`
	}{
		RefreshToken: refreshToken,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"code": 1, "message": "internal server error"}]}`))
		return
	}
}
