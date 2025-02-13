package email_confirmation_rest_adapter

import (
	"encoding/json"
	"errors"
	"net/http"

	email_confirmer "github.com/bratushkadan/floral/internal/auth/adapters/secondary/email/confirmer"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/internal/auth/service"
	"go.uber.org/zap"
)

type HandlerConfirmEmailRequestBody struct {
	Token string `json:"token"`
}
type HandlerSendConfirmation struct {
	Email string `json:"email"`
}

type HandlerResponseSuccess struct {
	Ok bool `json:"ok"`
}
type HandlerResponseFailure struct {
	Errors []string `json:"errors"`
}

type Adapter struct {
	l   *zap.Logger
	svc *service.EmailConfirmation
}

func New(svc *service.EmailConfirmation, l *zap.Logger) *Adapter {
	return &Adapter{
		l:   l,
		svc: svc,
	}
}

func (s *Adapter) HandleConfirmEmail(w http.ResponseWriter, r *http.Request) {
	var errs []string

	var b HandlerConfirmEmailRequestBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		errs = append(errs, "bad request body, 'token' field required.")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(&HandlerResponseFailure{Errors: errs}); err != nil {
			s.l.Error("failed to serialize response", zap.Error(err))
		}
		return
	}

	ctx := r.Context()
	if err := s.svc.Confirm(ctx, b.Token); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidConfirmationToken) || errors.Is(err, domain.ErrConfirmationTokenExpired):
			w.WriteHeader(http.StatusBadRequest)
			errs = append(errs, err.Error())
		default:
			s.l.Error("failed to confirm email", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			errs = append(errs, "failed to confirm email")
		}

		if err := json.NewEncoder(w).Encode(&HandlerResponseFailure{Errors: errs}); err != nil {
			s.l.Error("failed to serialize response", zap.Error(err))
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

func (s *Adapter) HandleSendConfirmation(w http.ResponseWriter, r *http.Request) {
	var errs []string

	var b HandlerSendConfirmation
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errs = append(errs, "bad request body, 'email' field required.")
		if err := json.NewEncoder(w).Encode(&HandlerResponseFailure{Errors: errs}); err != nil {
			s.l.Error("failed to serialize error response", zap.Error(err))
		}
		return
	}

	ctx := r.Context()
	if r.Host != "" {
		ctx = email_confirmer.ContextWithEmailConfirmationHost(ctx, r.Host)
	}

	if err := s.svc.Send(ctx, b.Email); err != nil {
		s.l.Error("failed to send confirmation email", zap.Error(err), zap.String("email", b.Email))
		w.WriteHeader(http.StatusInternalServerError)
		errs = append(errs, "failed to send confirmation email")
		if err := json.NewEncoder(w).Encode(&HandlerResponseFailure{Errors: errs}); err != nil {
			s.l.Error("failed to serialize error response", zap.Error(err))
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}
