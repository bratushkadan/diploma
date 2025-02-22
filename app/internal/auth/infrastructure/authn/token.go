package authn

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bratushkadan/floral/internal/auth/core/domain"
	"github.com/bratushkadan/floral/pkg/auth"
	"github.com/bratushkadan/floral/pkg/shared/api"
	"github.com/golang-jwt/jwt/v5"
)

const (
	RefreshTokenIdPrefix = "ry"
)

type RefreshTokenJwtClaims struct {
	TokenId   string           `json:"token_id"`
	SubjectId string           `json:"subject_id"`
	TokenType domain.TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type AccessTokenJwtClaims = api.AccessTokenJwtClaims

type TokenProvider struct {
	jwt *auth.JwtProvider
}

var _ domain.TokenProvider = (*TokenProvider)(nil)

type TokenProviderBuilder struct {
	privateKeyPath string
	publicKeyPath  string

	privateKey []byte
	publicKey  []byte

	p *TokenProvider
}

func NewTokenProviderBuilder() *TokenProviderBuilder {
	return &TokenProviderBuilder{p: &TokenProvider{}}
}

func (b *TokenProviderBuilder) PrivateKey(key []byte) *TokenProviderBuilder {
	b.privateKey = key
	return b
}
func (b *TokenProviderBuilder) PrivateKeyString(key string) *TokenProviderBuilder {
	b.privateKey = []byte(key)
	return b
}
func (b *TokenProviderBuilder) PublicKey(key []byte) *TokenProviderBuilder {
	b.publicKey = key
	return b
}
func (b *TokenProviderBuilder) PublicKeyString(key string) *TokenProviderBuilder {
	b.publicKey = []byte(key)
	return b
}
func (b *TokenProviderBuilder) PrivateKeyPath(path string) *TokenProviderBuilder {
	b.privateKeyPath = path
	return b
}
func (b *TokenProviderBuilder) PublicKeyPath(path string) *TokenProviderBuilder {
	b.publicKeyPath = path
	return b
}

func (b *TokenProviderBuilder) Build() (*TokenProvider, error) {
	var (
		privateKey, publicKey []byte
		err                   error
	)

	if len(b.privateKey) == 0 {
		if b.privateKeyPath == "" {
			return nil, errors.New("either a private key or a private key path must be provided to the token provider builder")
		}
		privateKey, err = os.ReadFile(b.privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key from file: %w", err)
		}
	} else {
		privateKey = b.privateKey
	}

	if len(b.publicKey) == 0 {
		if b.publicKeyPath == "" {
			return nil, errors.New("either a public key or a public key path must be provided to the token provider builder")
		}
		publicKey, err = os.ReadFile(b.publicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key from file: %w", err)
		}
	} else {
		publicKey = b.publicKey
	}

	jwtProvider, err := auth.NewJwtProviderBuilder().
		WithPrivateKey(privateKey).
		WithPublicKey(publicKey).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build jwt provider: %w", err)
	}

	b.p.jwt = jwtProvider

	return b.p, nil
}

func (p *TokenProvider) EncodeRefresh(token domain.RefreshToken) (string, error) {
	id := RefreshTokenIdPrefix + token.Id

	claims := RefreshTokenJwtClaims{
		TokenId:   id,
		SubjectId: token.SubjectId,
		TokenType: domain.TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(token.ExpiresAt),
		},
	}

	tokenString, err := p.jwt.Create(claims)
	if err != nil {
		return "", fmt.Errorf("failed to create jwt refresh token: %w", err)
	}
	return tokenString, nil
}

func (p *TokenProvider) DecodeRefresh(tokenString string) (domain.RefreshToken, error) {
	var claims RefreshTokenJwtClaims
	if err := p.jwt.Parse(tokenString, &claims); err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return domain.RefreshToken{
				Id:        claims.TokenId,
				SubjectId: claims.SubjectId,
				ExpiresAt: claims.RegisteredClaims.ExpiresAt.Time,
			}, domain.ErrTokenExpired
		}
		return domain.RefreshToken{}, fmt.Errorf("failed to parse jwt: %w: %w", err, domain.ErrTokenParseFailed)
	}

	if claims.TokenType != domain.TokenTypeRefresh {
		return domain.RefreshToken{}, fmt.Errorf(`expected token type to be "%s": %w`, domain.TokenTypeRefresh, domain.ErrInvalidTokenType)
	}

	id, found := strings.CutPrefix(claims.TokenId, RefreshTokenIdPrefix)
	if !found {
		return domain.RefreshToken{}, domain.ErrInvalidRefreshToken
	}

	return domain.RefreshToken{
		Id:        id,
		SubjectId: claims.SubjectId,
		ExpiresAt: claims.RegisteredClaims.ExpiresAt.Time,
	}, nil
}

func (p *TokenProvider) EncodeAccess(token domain.AccessToken) (string, error) {
	claims := AccessTokenJwtClaims{
		SubjectId:   token.SubjectId,
		SubjectType: token.SubjectType,
		TokenType:   domain.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(token.ExpiresAt),
		},
	}

	tokenString, err := p.jwt.Create(claims)
	if err != nil {
		return "", fmt.Errorf("failed to create jwt refresh token: %w", err)
	}

	return tokenString, nil
}

func (p *TokenProvider) DecodeAccess(tokenString string) (domain.AccessToken, error) {
	var claims AccessTokenJwtClaims
	if err := p.jwt.Parse(tokenString, &claims); err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return domain.AccessToken{}, domain.ErrTokenExpired
		}
		return domain.AccessToken{}, fmt.Errorf("failed to parse jwt: %w: %w", err, domain.ErrInvalidAccessToken)
	}

	if claims.TokenType != domain.TokenTypeAccess {
		return domain.AccessToken{}, fmt.Errorf(`expected token type to be "%s": %w`, domain.TokenTypeAccess, domain.ErrInvalidTokenType)
	}

	return domain.AccessToken{
		SubjectId:   claims.SubjectId,
		SubjectType: claims.SubjectType,
	}, nil
}
