package domain

import "context"

type AccountCreationNotificationProvider interface {
	Send(context.Context, SendAccountCreationNotificationDTOInput) (SendAccountCreationNotificationDTOOutput, error)
	// RcvProcess(context.Context, func(RcvProcessAccountCreationNotificationDTOInput) error) (RcvProcessAccountCreationNotificationDTOOutput, error)
}

type SendAccountCreationNotificationDTOInput struct {
	Email string
}
type SendAccountCreationNotificationDTOOutput struct {
}

// type RcvProcessAccountCreationNotificationDTOInput struct {
// 	Name  string
// 	Email string
// }
// type RcvProcessAccountCreationNotificationDTOOutput struct{}

type EmailConfirmationsNotificationProvider interface {
	Send(context.Context, SendEmailConfirmationNotificationsDTOInput) (SendEmailConfirmationNotificationsDTOOutput, error)
	// RcvProcess(context.Context, func(RcvProcessEmailConfirmationNotificationsDTOInput) error) (RcvProcessEmailConfirmationNotificationsDTOOutput, error)
}

type SendEmailConfirmationNotificationsDTOInput struct {
	Email string
}
type SendEmailConfirmationNotificationsDTOOutput struct {
}

// type RcvProcessEmailConfirmationNotificationsDTOInput struct {
// 	Name  string
// 	Email string
// }
// type RcvProcessEmailConfirmationNotificationsDTOOutput struct{}
