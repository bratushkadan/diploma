package authn

import (
	"fmt"
	"os"
	"time"

	"github.com/bratushkadan/floral/internal/auth/domain"
	"github.com/bratushkadan/floral/pkg/auth"
	"github.com/bratushkadan/floral/pkg/resource"
	"github.com/golang-jwt/jwt/v5"
)

const (
	RefreshTokenIdByteLength = 24
	RefreshTokenIdPrefix     = "ry"
)

type JwtProviderConf struct {
	PrivateKeyPath string
	PublicKeyPath  string
}

type JwtProvider struct {
	p *auth.JwtProvider
}

type RefreshTokenJwtProvider struct {
	jwt           *JwtProvider
	tokenDuration time.Duration
}
type AccessTokenJwtProvider struct {
	jwt           *JwtProvider
	tokenDuration time.Duration
}

var _ domain.RefreshTokenProvider = (*RefreshTokenJwtProvider)(nil)
var _ domain.AccessTokenProvider = (*AccessTokenJwtProvider)(nil)

func NewRefreshTokenJwtProvider(p *JwtProvider, tokenDuration time.Duration) *RefreshTokenJwtProvider {
	return &RefreshTokenJwtProvider{
		jwt:           p,
		tokenDuration: tokenDuration,
	}
}

func NewAccessTokenJwtProvider(p *JwtProvider, tokenDuration time.Duration) *AccessTokenJwtProvider {
	return &AccessTokenJwtProvider{
		jwt:           p,
		tokenDuration: tokenDuration,
	}
}

func (p *RefreshTokenJwtProvider) Create(subjectId string) (*domain.RefreshToken, string, error) {
	id := resource.GenerateIdPrefix(RefreshTokenIdByteLength, RefreshTokenIdPrefix)
	claims := auth.NewRefreshTokenJwtClaims(id, subjectId)
	claims.RegisteredClaims.Issuer = auth.FloralJwtIssuer
	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(p.tokenDuration))

	tokenString, err := p.jwt.p.Create(claims)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create jwt refresh token: %w", err)
	}

	return &domain.RefreshToken{
		TokenType: claims.TokenType,
		SubjectId: claims.SubjectId,
		ExpiresAt: claims.ExpiresAt.Time,
	}, tokenString, nil
}

func (p *RefreshTokenJwtProvider) Decode(tokenString string) (*domain.RefreshToken, error) {
	var claims auth.RefreshTokenJwtClaims
	if err := p.jwt.p.Parse(tokenString, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse jwt: %w", err)
	}

	if claims.TokenType != auth.RefreshTokenType {
		return nil, fmt.Errorf(`expected token type to be "%s"`, auth.RefreshTokenType)
	}
	if claims.ExpiresAt.After(time.Now()) {
		return nil, fmt.Errorf("token is expired")
	}
	if err := resource.ValidateIdByteLenPrefix(claims.TokenId, RefreshTokenIdByteLength, RefreshTokenIdPrefix); err != nil {
		return nil, fmt.Errorf("failed to validate jwt id: %w", err)
	}

	return &domain.RefreshToken{
		TokenId:   claims.TokenId,
		TokenType: claims.TokenType,
		SubjectId: claims.SubjectId,
		ExpiresAt: claims.RegisteredClaims.ExpiresAt.Time,
	}, nil
}

func (p *AccessTokenJwtProvider) Create(subjectId string, subjectType string) (*domain.AccessToken, string, error) {
	claims := auth.NewAccessTokenJwtClaims(subjectId, subjectType)
	claims.RegisteredClaims.Issuer = auth.FloralJwtIssuer
	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(p.tokenDuration))

	tokenString, err := p.jwt.p.Create(claims)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create jwt refresh token: %w", err)
	}

	return &domain.AccessToken{
		TokenType:   claims.TokenType,
		SubjectId:   claims.SubjectId,
		SubjectType: claims.SubjectType,
		ExpiresAt:   claims.ExpiresAt.Time,
	}, tokenString, nil
}

func (p *AccessTokenJwtProvider) Decode(tokenString string) (*domain.AccessToken, error) {
	var claims auth.AccessTokenJwtClaims
	if err := p.jwt.p.Parse(tokenString, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse string to jwt token: %w", err)
	}

	if claims.TokenType != auth.AccessTokenType {
		return nil, fmt.Errorf(`expected token type to be "%s"`, auth.AccessTokenType)
	}
	if claims.ExpiresAt.After(time.Now()) {
		return nil, fmt.Errorf("token is expired")
	}

	return &domain.AccessToken{
		TokenType:   claims.TokenType,
		SubjectId:   claims.SubjectId,
		SubjectType: claims.SubjectId,
		ExpiresAt:   claims.RegisteredClaims.ExpiresAt.Time,
	}, nil
}

func NewJwtProvider(conf JwtProviderConf) (*JwtProvider, error) {
	privateKey, err := os.ReadFile(conf.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key from file: %w", err)
	}
	publicKey, err := os.ReadFile(conf.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key from file: %w", err)
	}

	jwtProvider, err := auth.NewJwtProviderBuilder().
		WithPrivateKey(privateKey).
		WithPublicKey(publicKey).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build jwt provider: %w", err)
	}

	return &JwtProvider{p: jwtProvider}, nil
}
