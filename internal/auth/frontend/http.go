package frontend

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/go-chi/chi/v5"
)

type HttpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *HttpError) Error() string {
	return e.Message
}

type HttpErrors struct {
	Errors []HttpError `json:"errors"`
}

func NewHttpErrors(errs ...HttpError) HttpErrors {
	return HttpErrors{
		Errors: errs,
	}
}

var (
	ErrHttpInternalServerError = HttpError{
		Code:    1,
		Message: "internal server error",
	}
	ErrHttpBadRequestBody = HttpError{
		Code:    2,
		Message: "bad request body",
	}
	ErrHttpInvalidCredentials = HttpError{
		Code:    3,
		Message: "bad user credentials",
	}
	ErrHttpInvalidRefreshToken = HttpError{
		Code:    4,
		Message: "invalid refresh token",
	}
	ErrHttpInvalidAccessToken = HttpError{
		Code:    5,
		Message: "invalid access token",
	}
	ErrHttpAccessDenied = HttpError{
		Code:    6,
		Message: "access denied",
	}
	NewErrHttpEmailIsInUse = func(email string) HttpError {
		return HttpError{
			Code:    7,
			Message: fmt.Sprintf(`email "%s" is already in use`, email),
		}
	}
	ErrHttpEmailIsNotConfirmed = HttpError{
		Code:    8,
		Message: "email is not confirmed",
	}
	ErrHttpBadEmailConfirmationId = HttpError{
		Code:    9,
		Message: "bad email confirmation id",
	}
)

type HttpImpl struct {
	auth                        *domain.AuthService
	confirmationIdQueryParamKey string
}

func NewHttpImpl(auth *domain.AuthService, confirmationIdQueryParamKey string) *HttpImpl {
	return &HttpImpl{
		auth:                        auth,
		confirmationIdQueryParamKey: confirmationIdQueryParamKey,
	}
}

func (f *HttpImpl) Start(auth *domain.AuthService) error {
	f.auth = auth

	r := chi.NewRouter()

	apiRouter := chi.NewRouter()

	v1Router := chi.NewRouter()

	apiRouter.Mount("/v1", v1Router)

	usersRouter := chi.NewRouter()
	usersConfirmationRouter := chi.NewRouter()

	v1Router.Mount("/users", usersRouter)
	v1Router.Mount("/usersConfirmation", usersConfirmationRouter)

	usersRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("i am a teapot"))
	})
	usersRouter.Post("/:register", http.HandlerFunc(f.RegisterUserHandler))
	usersRouter.Post("/:registerSeller", http.HandlerFunc(f.RegisterSellerHandler))
	usersRouter.Post("/:registerAdmin", http.HandlerFunc(f.RegisterAdminHandler))
	usersRouter.Post("/:authenticate", http.HandlerFunc(f.AuthenticateHandler))
	usersRouter.Post("/:renewRefreshToken", http.HandlerFunc(f.RenewRefreshTokenHandler))
	usersRouter.Post("/:createAccessToken", http.HandlerFunc(f.CreateAccessToken))

	usersConfirmationRouter.Get("/:confirmEmail", http.HandlerFunc(f.ConfirmEmailHandler))

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
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			log.Print(err)
		}
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
		if errors.Is(err, domain.ErrEmailIsInUse) {
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(NewErrHttpEmailIsInUse(reqData.Email))); err != nil {
				log.Print(err)
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
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
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
		return
	}
}

func (f *HttpImpl) RegisterAdminHandler(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		Admin struct {
			Name     string `json:"name"`
			Password string `json:"password"`
			Email    string `json:"email"`
		} `json:"admin"`
		SecretToken string `json:"secret_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			log.Print(err)
		}
		return
	}

	if reqData.SecretToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			log.Print(err)
		}
		return
	}

	user, err := f.auth.CreateAdmin(r.Context(), domain.CreateAdminReq{
		CreateUserReq: domain.CreateUserReq{
			Name:     reqData.Admin.Name,
			Password: reqData.Admin.Password,
			Email:    reqData.Admin.Email,
		},
		SecretToken: reqData.SecretToken,
	})
	if err != nil {
		log.Print(err)
		if errors.Is(err, domain.ErrEmailIsInUse) {
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(NewErrHttpEmailIsInUse(reqData.Admin.Email))); err != nil {
				log.Print(err)
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
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
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
		return
	}
}

func (f *HttpImpl) RegisterSellerHandler(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		Seller struct {
			Name     string `json:"name"`
			Password string `json:"password"`
			Email    string `json:"email"`
		} `json:"seller"`
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			log.Print(err)
		}
		return
	}

	user, err := f.auth.CreateSeller(r.Context(), domain.CreateSellerReq{
		CreateUserReq: domain.CreateUserReq{Name: reqData.Seller.Name,
			Password: reqData.Seller.Password,
			Email:    reqData.Seller.Email,
		},
	}, reqData.AccessToken)
	if err != nil {
		log.Print(err)
		// FIXME: domain ничего не должен знать про access/refresh-токены: нужно понять, на каком уровне нужно организовать security, валидацию токенов и парсинг credentials из security-токенов
		if errors.Is(err, domain.ErrInvalidAccessToken) {
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInvalidAccessToken)); err != nil {
				log.Print(err)
			}
			return
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpAccessDenied)); err != nil {
				log.Print(err)
			}
			return
		}
		if errors.Is(err, domain.ErrEmailIsInUse) {
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(NewErrHttpEmailIsInUse(reqData.Seller.Email))); err != nil {
				log.Print(err)
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
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
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
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
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			log.Print(err)
		}
		return
	}

	refreshToken, err := f.auth.Authenticate(r.Context(), reqData.Email, reqData.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
				log.Print(err)
			}
			return
		}
		if errors.Is(err, domain.ErrAccountEmailNotConfirmed) {
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpEmailIsNotConfirmed)); err != nil {
				log.Print(err)
			}
			return
		}
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(&struct {
		RefreshToken string `json:"refresh_token"`
	}{
		RefreshToken: refreshToken,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
		return
	}
}

func (f *HttpImpl) RenewRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil || reqData.RefreshToken == "" {
		log.Print(err)
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			log.Print(err)
		}
		return
	}

	refreshToken, err := f.auth.RenewRefreshToken(r.Context(), reqData.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInvalidRefreshToken)); err != nil {
				log.Print(err)
			}
			return
		}
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(&struct {
		RefreshToken string `json:"refresh_token"`
	}{
		RefreshToken: refreshToken,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
		return
	}
}

func (f *HttpImpl) CreateAccessToken(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil || reqData.RefreshToken == "" {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			log.Print(err)
		}
		return
	}

	accessToken, accessTokenStr, err := f.auth.GetAccessToken(r.Context(), reqData.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAccessToken) {
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInvalidRefreshToken)); err != nil {
				log.Print(err)
			}
			return
		}
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(&struct {
		AccessToken string    `json:"access_token"`
		ExpiresAt   time.Time `json:"expires_at"`
	}{
		AccessToken: accessTokenStr,
		ExpiresAt:   accessToken.ExpiresAt,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
	}
}

func (f *HttpImpl) ConfirmEmailHandler(w http.ResponseWriter, r *http.Request) {
	confirmationId := r.URL.Query().Get(f.confirmationIdQueryParamKey)

	if confirmationId == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadEmailConfirmationId)); err != nil {
			log.Print(err)
		}
		return
	}

	if err := f.auth.ConfirmEmail(r.Context(), confirmationId); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			log.Print(err)
		}
		return
	}

	w.Write([]byte("Email is successfully confirmed."))
}
