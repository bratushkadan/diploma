package auth

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	RefreshTokenType = "refresh"
	AccessTokenType  = "access"
)

var (
	FloralJwtIssuer = "authorization.floral.io"
)

type RefreshTokenJwtClaims struct {
	TokenType string `json:"token_type"`
	SubjectId string `json:"subject_id"`
	jwt.RegisteredClaims
}

func NewRefreshTokenJwtClaims(subjectId string) *RefreshTokenJwtClaims {
	claims := &RefreshTokenJwtClaims{
		TokenType: RefreshTokenType,
		SubjectId: subjectId,
	}
	claims.RegisteredClaims.Issuer = FloralJwtIssuer
	return claims
}

type AccessTokenJwtClaims struct {
	TokenType   string `json:"token_type"`
	SubjectId   string `json:"subject_id"`
	SubjectType string `json:"subject_type"`
	jwt.RegisteredClaims
}

func NewAccessTokenJwtClaims(subjectId string, subjectType string) *AccessTokenJwtClaims {
	claims := &AccessTokenJwtClaims{
		TokenType:   AccessTokenType,
		SubjectId:   subjectId,
		SubjectType: subjectType,
	}
	claims.RegisteredClaims.Issuer = FloralJwtIssuer
	return claims
}

type JwtProvider struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func (p *JwtProvider) Create(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(p.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return tokenString, nil
}

func (p *JwtProvider) Parse(token string, claims jwt.Claims) error {
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method for token")
		}
		return p.publicKey, nil
	})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}
	if !parsedToken.Valid {
		return fmt.Errorf("invalid refresh token")
	}
	issuer, err := claims.GetIssuer()
	if err != nil {
		return fmt.Errorf(`failed to get issuer of jwt, expected "%s"`, FloralJwtIssuer)
	}
	if issuer != FloralJwtIssuer {
		return fmt.Errorf(`invalid refresh token issuer, expected "%s"`, FloralJwtIssuer)
	}

	return nil
}

// FIXME: fix the issue of private key being required
type JwtProviderConf struct {
	PrivateKey []byte
	PublicKey  []byte
}

func NewJwtProvider(conf JwtProviderConf) (*JwtProvider, error) {
	if len(conf.PrivateKey) == 0 {
		return nil, fmt.Errorf("private key can't be empty bytes")
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(conf.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key from PEM for asymmetric JWT signing: %w", err)
	}
	if len(conf.PublicKey) == 0 {
		return nil, fmt.Errorf("public key can't be empty bytes")
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(conf.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key from PEM for asymmetric JWT validation: %w", err)
	}

	return &JwtProvider{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}
