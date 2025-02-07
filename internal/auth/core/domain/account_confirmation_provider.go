package domain

import "context"

type AccountConfirmationProvider interface {
	Send(context.Context, SendAccountConfirmationDTOInput) (SendAccountConfirmationDTOOutput, error)
}

type SendAccountConfirmationDTOInput struct {
	Name  string
	Email string
}
type SendAccountConfirmationDTOOutput struct {
}
