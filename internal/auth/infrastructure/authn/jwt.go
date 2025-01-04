package authn

import (
	"fmt"
	"os"
	"time"

	"github.com/bratushkadan/floral/internal/auth/domain"
	"github.com/bratushkadan/floral/pkg/auth"
	"github.com/golang-jwt/jwt/v5"
)

type JwtProviderConf struct {
	privateKeyPath string
	publicKeyPath  string
}

type JwtProvider struct {
	p *auth.JwtProvider
}

type RefreshTokenJwtProvider struct {
	prov          *auth.JwtProvider
	tokenDuration time.Duration
}
type AccessTokenJwtProvider struct {
	prov          *auth.JwtProvider
	tokenDuration time.Duration
}

var _ domain.RefreshTokenProvider = (*RefreshTokenJwtProvider)(nil)
var _ domain.AccessTokenProvider = (*AccessTokenJwtProvider)(nil)

func NewRefreshTokenJwtProvider(p *JwtProvider, tokenDuration time.Duration) *RefreshTokenJwtProvider {
	return &RefreshTokenJwtProvider{
		prov:          p.p,
		tokenDuration: tokenDuration,
	}
}

func NewAccessTokenJwtProvider(p *JwtProvider, tokenDuration time.Duration) *AccessTokenJwtProvider {
	return &AccessTokenJwtProvider{
		prov:          p.p,
		tokenDuration: tokenDuration,
	}
}

func (p *RefreshTokenJwtProvider) Create(subjectId string) (*domain.RefreshToken, string, error) {
	claims := auth.NewRefreshTokenJwtClaims(subjectId)
	claims.RegisteredClaims.Issuer = auth.FloralJwtIssuer
	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(p.tokenDuration))

	tokenString, err := p.prov.Create(claims)
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
	if err := p.prov.Parse(tokenString, &claims); err != nil {
		return nil, fmt.Errorf("")
	}

	return &domain.RefreshToken{
		TokenType: claims.TokenType,
		SubjectId: claims.SubjectId,
		ExpiresAt: claims.RegisteredClaims.ExpiresAt.Time,
	}, nil
}

func (p *AccessTokenJwtProvider) Create(subjectId string, subjectType string) (*domain.AccessToken, string, error) {
	claims := auth.NewAccessTokenJwtClaims(subjectId, subjectType)
	claims.RegisteredClaims.Issuer = auth.FloralJwtIssuer
	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(p.tokenDuration))

	tokenString, err := p.prov.Create(claims)
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
	if err := p.prov.Parse(tokenString, &claims); err != nil {
		return nil, fmt.Errorf("")
	}

	return &domain.AccessToken{
		TokenType:   claims.TokenType,
		SubjectId:   claims.SubjectId,
		SubjectType: claims.SubjectId,
		ExpiresAt:   claims.RegisteredClaims.ExpiresAt.Time,
	}, nil
}

func NewJwtProvider(conf JwtProviderConf) (*JwtProvider, error) {
	privateKey, err := os.ReadFile(conf.privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key from file: %w", err)
	}
	publicKey, err := os.ReadFile(conf.publicKeyPath)
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
