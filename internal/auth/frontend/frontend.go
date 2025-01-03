package frontend

import "github.com/bratushkadan/floral/internal/auth/domain"

type FrontEnd interface {
	Start(auth *domain.AuthService) error
}
