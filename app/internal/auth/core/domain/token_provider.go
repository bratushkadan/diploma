package domain

type TokenProvider interface {
	EncodeRefresh(token RefreshToken) (tokenString string, err error)
	DecodeRefresh(token string) (RefreshToken, error)
	EncodeAccess(token AccessToken) (tokenString string, err error)
	DecodeAccess(token string) (AccessToken, error)
}
