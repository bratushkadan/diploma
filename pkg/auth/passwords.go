package auth

import "golang.org/x/crypto/bcrypt"

type PasswordHasher struct {
	pepper string
}

func NewPasswordHasher(secretPhrase string) *PasswordHasher {
	return &PasswordHasher{
		pepper: secretPhrase,
	}
}

func (h *PasswordHasher) Hash(pass string) (string, error) {
	pass = pass + h.pepper
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPass), nil
}
func (h *PasswordHasher) Check(pass string, hashedPass string) bool {
	pass = pass + h.pepper
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(pass)); err != nil {
		return false
	}
	return true
}
