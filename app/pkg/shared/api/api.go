package api

type AccountCreationMessage struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

type AccountConfirmationMessage struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}
