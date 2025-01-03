package auth_test

import (
	_ "embed"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bratushkadan/floral/pkg/auth"
	"github.com/golang-jwt/jwt/v5"
)

//go:embed test_fixtures/private.key
var privateKey []byte

//go:embed test_fixtures/public.key
var publicKey []byte

func TestJwt(t *testing.T) {
	conf := auth.JwtProviderConf{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
	tokenProv, err := auth.NewJwtProvider(conf)
	if err != nil {
		t.Fatal(err)
	}

	refreshTokenSubjectId := "dan"
	claims := auth.NewRefreshTokenJwtClaims(refreshTokenSubjectId)

	tokenString, err := tokenProv.Create(claims)
	if err != nil {
		t.Fatal(err)
	}

	var decodedClaims auth.RefreshTokenJwtClaims
	tokenProv.Parse(tokenString, &decodedClaims)

	if decodedClaims.TokenType != auth.RefreshTokenType {
		fmt.Errorf(`expected TokenType of refresh token decodedClaims to be "%s", got "%s"`, auth.RefreshTokenType, decodedClaims.TokenType)
	}
	if decodedClaims.SubjectId != refreshTokenSubjectId {
		fmt.Errorf(`expected SubjectId of refresh token decodedClaims to be "%s", got "%s"`, refreshTokenSubjectId, decodedClaims.TokenType)
	}
}

func TestJwtDecodeExpired(t *testing.T) {
	conf := auth.JwtProviderConf{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
	tokenProv, err := auth.NewJwtProvider(conf)
	if err != nil {
		t.Fatal(err)
	}

	refreshTokenSubjectId := "dan"
	claims := auth.NewRefreshTokenJwtClaims(refreshTokenSubjectId)
	claims.RegisteredClaims.ExpiresAt = &jwt.NumericDate{Time: time.Now().Add(-1 * time.Minute)}

	tokenString, err := tokenProv.Create(claims)
	if err != nil {
		t.Fatal(err)
	}

	var decodedClaims auth.RefreshTokenJwtClaims
	err = tokenProv.Parse(tokenString, &decodedClaims)
	if err == nil {
		t.Fatal("expected error while decoding token with the negative token duration value")
	}
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("unexpected token provider decoding error: %v", err)
	}
}
