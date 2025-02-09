package auth_test

import (
	_ "embed"
	"testing"
	"time"

	"github.com/bratushkadan/floral/pkg/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

//go:embed test_fixtures/private.key
var privateKey []byte

//go:embed test_fixtures/public.key
var publicKey []byte

type testClaims struct {
	SubjectId string `json:"subject_id"`
	jwt.RegisteredClaims
}

func getJwtProvider() (*auth.JwtProvider, error) {
	return auth.NewJwtProviderBuilder().
		WithPublicKey(publicKey).
		WithPrivateKey(privateKey).
		Build()
}

func TestJwt(t *testing.T) {
	tokenProv, err := getJwtProvider()
	assert.NoError(t, err)

	subjectId := "dan"
	claims := testClaims{
		SubjectId: subjectId,
	}

	tokenString, err := tokenProv.Create(claims)
	assert.NoError(t, err)

	var decodedClaims testClaims
	tokenProv.Parse(tokenString, &decodedClaims)

	assert.Equal(t, subjectId, decodedClaims.SubjectId, "encoded and decoded jwt subject ids should be equal")
}

func TestJwtDecodeExpired(t *testing.T) {
	tokenProv, err := getJwtProvider()
	assert.NoError(t, err)

	subjectId := "dan"
	claims := testClaims{
		SubjectId: subjectId,
	}
	claims.RegisteredClaims.ExpiresAt = &jwt.NumericDate{Time: time.Now().Add(-1 * time.Minute)}

	tokenString, err := tokenProv.Create(claims)
	assert.NoError(t, err)

	var decodedClaims testClaims
	err = tokenProv.Parse(tokenString, &decodedClaims)

	if assert.NotNil(t, err) {
		assert.ErrorIs(t, err, jwt.ErrTokenExpired, "unexpected token provider decoding error")
	}
}
