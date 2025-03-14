package auth

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JwtProvider struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	parserOpts []jwt.ParserOption
}

type JwtProviderBuilder struct {
	privateKey []byte
	publicKey  []byte
	prov       JwtProvider
}

func NewJwtProviderBuilder() *JwtProviderBuilder {
	return &JwtProviderBuilder{prov: JwtProvider{}}
}

func (b *JwtProviderBuilder) WithPrivateKey(privateKey []byte) *JwtProviderBuilder {
	b.privateKey = privateKey
	return b
}
func (b *JwtProviderBuilder) WithPublicKey(publicKey []byte) *JwtProviderBuilder {
	b.publicKey = publicKey
	return b
}

// Example: jwt.WithIssuer("foo")
func (b *JwtProviderBuilder) WithParserOptions(opts ...jwt.ParserOption) *JwtProviderBuilder {
	b.prov.parserOpts = opts
	return b
}

func (b *JwtProviderBuilder) Build() (*JwtProvider, error) {
	if len(b.publicKey) == 0 {
		return nil, fmt.Errorf("public key can't be empty")
	}
	publicKey, err := jwt.ParseECPublicKeyFromPEM(b.publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ECC public key from PEM for asymmetric JWT validation: %w", err)
	}

	var privateKey *ecdsa.PrivateKey
	if b.privateKey != nil {
		if len(b.privateKey) == 0 {
			return nil, fmt.Errorf("private key can't be empty bytes")
		}
		var err error
		privateKey, err = jwt.ParseECPrivateKeyFromPEM(b.privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ECC private key from PEM for asymmetric JWT signing: %w", err)
		}
	}

	b.prov.publicKey = publicKey
	b.prov.privateKey = privateKey

	return &b.prov, nil
}

func (p *JwtProvider) Create(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	tokenString, err := token.SignedString(p.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return tokenString, nil
}

func (p *JwtProvider) Parse(token string, claims jwt.Claims) error {
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method for token")
		}
		return p.publicKey, nil
	}, p.parserOpts...)
	if err != nil {
		return fmt.Errorf("failed to parse jwt token: %w", err)
	}
	if !parsedToken.Valid {
		return fmt.Errorf("invalid jwt token")
	}

	return nil
}
