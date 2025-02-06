package frontend

import "github.com/bratushkadan/floral/internal/auth/service"

type FrontEnd interface {
	Start(auth *service.Auth) error
}
