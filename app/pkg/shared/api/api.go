package api

type AccountCreationMessage struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

// TODO: send id only, decouple account activation from email from email
type AccountConfirmationMessage struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}
