package api

type AccountCreationMessage struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AccountConfirmationMessage struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
