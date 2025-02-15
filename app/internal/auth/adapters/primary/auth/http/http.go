package http_adapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
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
	ErrHttpRefreshTokenToReplaceNotFound = HttpError{
		Code:    10,
		Message: "refresh token to replace not found",
	}
)

type Http struct {
	svc domain.AuthService
	l   *zap.Logger

	validateJson *validator.Validate
}

type HttpBuilder struct {
	http Http
}

func NewBuilder() *HttpBuilder {
	return &HttpBuilder{}
}

func (b *HttpBuilder) Svc(svc domain.AuthService) *HttpBuilder {
	b.http.svc = svc
	return b
}
func (b *HttpBuilder) Logger(l *zap.Logger) *HttpBuilder {
	b.http.l = l
	return b
}

func (b *HttpBuilder) Build() (*Http, error) {
	if b.http.svc == nil {
		return nil, errors.New("auth service must be set for http builder")
	}

	if b.http.l == nil {
		b.http.l = zap.NewNop()
	}

	b.http.validateJson = validator.New(validator.WithRequiredStructEnabled())

	return &b.http, nil
}

type RegisterUserHandlerReq struct {
	Name     string `json:"name" validate:"required,min=2,max=40"`
	Password string `json:"password" validate:"required,min=8,max=24"`
	Email    string `json:"email" validate:"required,email"`
}
type RegisterUserHandlerRes struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (f *Http) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var reqData RegisterUserHandlerReq
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		f.l.Info("failed to decode request body", zap.String("handler", "RegisterUser"), zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}
	if err := f.validateJson.Struct(reqData); err != nil {
		f.l.Info("invalid request struct", zap.String("handler", "RegisterUser"), zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}

	user, err := f.svc.CreateUser(r.Context(), domain.CreateUserReq{
		Name:     reqData.Name,
		Password: reqData.Password,
		Email:    reqData.Email,
	},
	)
	if err != nil {
		if errors.Is(err, domain.ErrEmailIsInUse) {
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(NewErrHttpEmailIsInUse(reqData.Email))); err != nil {
				f.l.Error("failed to encode error response", zap.Error(err))
			}
			return
		}
		f.l.Error("unexpected error occurred in handler RegisterUserHandler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("failed to encode internal server error response", zap.Error(err))
		}
		return
	}

	if err := json.NewEncoder(w).Encode(&RegisterUserHandlerRes{
		Id:   user.Id,
		Name: user.Name,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("failed to encode internal server error response", zap.Error(err))
		}
		return
	}
}

type RegisterAdminHandlerReq struct {
	Name     string `json:"name" validate:"required,min=2,max=40"`
	Password string `json:"password" validate:"required,min=8,max=24"`
	Email    string `json:"email" validate:"required,email"`
}
type RegisterAdminHandlerRes struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (f *Http) RegisterAdminHandler(w http.ResponseWriter, r *http.Request) {
	var reqData RegisterAdminHandlerReq
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		f.l.Info("failed to decode request body for handler RegisterAdminHandler", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}
	if err := f.validateJson.Struct(reqData); err != nil {
		f.l.Info("invalid request struct", zap.String("handler", "RegisterAdminHandler"), zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}

	user, err := f.svc.CreateAdmin(r.Context(), domain.CreateAdminReq{
		Name:     reqData.Name,
		Password: reqData.Password,
		Email:    reqData.Email,
	})
	if err != nil {
		if errors.Is(err, domain.ErrEmailIsInUse) {
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(NewErrHttpEmailIsInUse(reqData.Email))); err != nil {
				f.l.Error("failed to encode error response", zap.Error(err))
			}
			return
		}
		f.l.Error("unexpected error occurred in handler RegisterAdminHandler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("unexpected error occurred in handler AuthenticateHandler", zap.Error(err))
		}
		return
	}

	if err := json.NewEncoder(w).Encode(&RegisterAdminHandlerRes{
		Id:   user.Id,
		Name: user.Name,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("unexpected error occurred in handler AuthenticateHandler", zap.Error(err))
		}
		return
	}
}

type RegisterSellerHandlerReq struct {
	Seller struct {
		Name     string `json:"name" validate:"required,min=2,max=40"`
		Password string `json:"password" validate:"required,min=8,max=24"`
		Email    string `json:"email" validate:"required,email"`
	} `json:"seller" validate:"required"`
	AccessToken string `json:"access_token" validate:"required"`
}
type RegisterSellerHandlerRes struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (f *Http) RegisterSellerHandler(w http.ResponseWriter, r *http.Request) {
	var reqData RegisterSellerHandlerReq
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		f.l.Info("failed to decode request body for handler RegisterSellerHandler", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}
	if err := f.validateJson.Struct(reqData); err != nil {
		f.l.Info("invalid request struct", zap.String("handler", "RegisterSellerHandler"), zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}

	user, err := f.svc.CreateSeller(r.Context(), domain.CreateSellerReq{
		Name:        reqData.Seller.Name,
		Password:    reqData.Seller.Password,
		Email:       reqData.Seller.Email,
		AccessToken: reqData.AccessToken,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAccessToken) {
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInvalidAccessToken)); err != nil {
				f.l.Error("failed to encode error response", zap.Error(err))
			}
			return
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpAccessDenied)); err != nil {
				f.l.Error("failed to encode error response", zap.Error(err))
			}
			return
		}
		if errors.Is(err, domain.ErrEmailIsInUse) {
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(NewErrHttpEmailIsInUse(reqData.Seller.Email))); err != nil {
				f.l.Error("failed to encode error response", zap.Error(err))
			}
			return
		}
		f.l.Error("unexpected error occurred in handler RegisterSellerHandler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("unexpected error occurred in handler AuthenticateHandler", zap.Error(err))
		}
		return
	}

	if err := json.NewEncoder(w).Encode(&RegisterSellerHandlerRes{
		Id:   user.Id,
		Name: user.Name,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("unexpected error occurred in handler AuthenticateHandler", zap.Error(err))
		}
		return
	}
}

type ActivateAccountHandlerReq struct {
	Email string `json:"email" validate:"required,email"`
}
type ActivateAccountHandlerRes struct {
	Ok bool `json:"ok"`
}

func (f *Http) ActivateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var reqData ActivateAccountHandlerReq
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		f.l.Info("failed to decode request body for handler ActivateAccountHandler", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}
	if err := f.validateJson.Struct(reqData); err != nil {
		f.l.Info("invalid request struct", zap.String("handler", "ActivateAccountHandler"), zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}

	_, err := f.svc.ActivateAccounts(r.Context(), domain.ActivateAccountsReq{
		Emails: []string{reqData.Email},
	})
	if err != nil {
		f.l.Error("unexpected error occurred in handler ActivateAccountHandler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("unexpected error occurred in handler ActivateAccountHandler", zap.Error(err))
		}
		return
	}

	if err := json.NewEncoder(w).Encode(&ActivateAccountHandlerRes{Ok: true}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("unexpected error occurred in handler ActivateAccountHandler", zap.Error(err))
		}
		return
	}
}

type AuthenticateHandlerReq struct {
	Password string `json:"password" validate:"required,min=8,max=24"`
	Email    string `json:"email" validate:"required,email"`
}
type AuthenticateHandlerRes struct {
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (f *Http) AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	var reqData AuthenticateHandlerReq
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		f.l.Info("failed to decode request body for handler AuthenticateHandler", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}
	if err := f.validateJson.Struct(reqData); err != nil {
		f.l.Info("invalid request struct", zap.String("handler", "AuthenticateHandler"), zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}

	res, err := f.svc.Authenticate(r.Context(), domain.AuthenticateReq{
		Email:    reqData.Email,
		Password: reqData.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
				f.l.Error("failed to encode error response", zap.Error(err))
			}
			return
		}
		if errors.Is(err, domain.ErrAccountNotActivated) {
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpEmailIsNotConfirmed)); err != nil {
				f.l.Error("failed to encode error response", zap.Error(err))
			}
			return
		}
		f.l.Error("unexpected error occurred in handler AuthenticateHandler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("failed to encode internal server error response", zap.Error(err))
		}
		return
	}

	if err := json.NewEncoder(w).Encode(&AuthenticateHandlerRes{
		RefreshToken: res.RefreshToken,
		ExpiresAt:    res.ExpiresAt,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("failed to encode internal server error response", zap.Error(err))
		}
		return
	}
}

type ReplaceRefreshTokenHandlerReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
type ReplaceRefreshTokenHandlerRes struct {
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (f *Http) ReplaceRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var reqData ReplaceRefreshTokenHandlerReq
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil || reqData.RefreshToken == "" {
		f.l.Info("failed to decode request body for handler CreateAccessToken", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}
	if err := f.validateJson.Struct(reqData); err != nil {
		f.l.Info("invalid request struct", zap.String("handler", "ReplaceRefreshTokenHandler"), zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}

	res, err := f.svc.ReplaceRefreshToken(r.Context(), domain.ReplaceRefreshTokenReq{
		RefreshToken: reqData.RefreshToken,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInvalidRefreshToken)); err != nil {
				f.l.Error("failed to encode error response", zap.Error(err))
			}
			return
		}
		if errors.Is(err, domain.ErrRefreshTokenToReplaceNotFound) {
			w.WriteHeader(http.StatusNotFound)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpRefreshTokenToReplaceNotFound)); err != nil {
				f.l.Error("failed to encode internal server error response", zap.Error(err))
			}
			return
		}
		f.l.Error("unexpected error occurred in handler ReplaceRefreshTokenHandler", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("failed to encode internal server error response", zap.Error(err))
		}
		return
	}

	if err := json.NewEncoder(w).Encode(&ReplaceRefreshTokenHandlerRes{
		RefreshToken: res.RefreshToken,
		ExpiresAt:    res.ExpiresAt,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("failed to encode internal server error response", zap.Error(err))
		}
		return
	}
}

type CreateAccessTokenReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
type CreateAccessTokenRes struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (f *Http) CreateAccessToken(w http.ResponseWriter, r *http.Request) {
	var reqData CreateAccessTokenReq
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil || reqData.RefreshToken == "" {
		f.l.Info("failed to decode request body for handler CreateAccessToken", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode internal server error response", zap.Error(err))
		}
		return
	}
	if err := f.validateJson.Struct(reqData); err != nil {
		f.l.Info("invalid request struct", zap.String("handler", "CreateAccessToken"), zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpBadRequestBody)); err != nil {
			f.l.Error("failed to encode error response", zap.Error(err))
		}
		return
	}

	res, err := f.svc.CreateAccessToken(r.Context(), domain.CreateAccessTokenReq{
		RefreshToken: reqData.RefreshToken,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAccessToken) {
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInvalidRefreshToken)); err != nil {
				f.l.Error("failed to encode internal server error response", zap.Error(err))
			}
			return
		}
		f.l.Error("unexpected error occurred in handler CreateAccessToken", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("failed to encode internal server error response", zap.Error(err))
		}
		return
	}

	if err := json.NewEncoder(w).Encode(&struct {
		AccessToken string    `json:"access_token"`
		ExpiresAt   time.Time `json:"expires_at"`
	}{
		AccessToken: res.AccessToken,
		ExpiresAt:   res.ExpiresAt,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(NewHttpErrors(ErrHttpInternalServerError)); err != nil {
			f.l.Error("failed to encode internal server error response", zap.Error(err))
		}
	}
}
