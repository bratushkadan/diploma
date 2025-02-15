package email_confirmation_http_adapter

import (
	"encoding/json"
	"errors"
	"net/http"

	email_confirmer "github.com/bratushkadan/floral/internal/auth/adapters/secondary/email/confirmer"
	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/shared/api"
	"github.com/bratushkadan/floral/pkg/yc/serverless/ymq"
	"go.uber.org/zap"
)

type HandlerConfirmEmailRequestBody struct {
	Token string `json:"token"`
}

type HandlerResponseSuccess struct {
	Ok bool `json:"ok"`
}
type HandlerResponseFailure struct {
	Errors []string `json:"errors"`
}

type Adapter struct {
	l   *zap.Logger
	svc domain.AccountEmailConfirmation
}

func New(svc domain.AccountEmailConfirmation, l *zap.Logger) *Adapter {
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

	var b api.AccountConfirmationMessage
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errs = append(errs, "bad request body, 'email' field is required.")
		if err := json.NewEncoder(w).Encode(&HandlerResponseFailure{Errors: errs}); err != nil {
			s.l.Error("failed to serialize error response", zap.Error(err))
		}
		return
	}

	ctx := r.Context()

	host := r.Header.Get("Host")
	if host == "" {
		host = r.Host
	}
	if host != "" {
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

func (s *Adapter) HandleSendConfirmationYmqTrigger(w http.ResponseWriter, r *http.Request) {
	var errs []string

	var reqBody ymq.YMQRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		s.l.Info("failed to decode ymq request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		errs = append(errs, "bad request body")
		if err := json.NewEncoder(w).Encode(&HandlerResponseFailure{Errors: errs}); err != nil {
			s.l.Error("failed to serialize error response", zap.Error(err))
		}
		return
	}

	ctx := r.Context()

	host := r.Header.Get("Host")
	if host == "" {
		host = r.Host
	}
	if host != "" {
		ctx = email_confirmer.ContextWithEmailConfirmationHost(ctx, r.Host)
	}

	// Messsage group size for email confirmation trigger is always 1.
	for _, msg := range reqBody.Messages {
		var b api.AccountCreationMessage
		if err := json.Unmarshal([]byte(msg.Details.Message.Body), &b); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			s.l.Error("bad message format in request body", zap.Error(err))
			errs = append(errs, "bad message format in request body, 'email' field is required.")
		}
		if err := json.NewEncoder(w).Encode(&HandlerResponseFailure{Errors: errs}); err != nil {
			s.l.Error("failed to serialize error response", zap.Error(err))
			return
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
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}
