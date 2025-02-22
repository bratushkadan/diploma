package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	jwt_auth "github.com/bratushkadan/floral/pkg/auth"
	"github.com/bratushkadan/floral/pkg/shared/api"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RequiredRoute struct {
	Method string
	Path   string
}

func NewRequiredRoute(method, path string) RequiredRoute {
	return RequiredRoute{
		Method: method,
		Path:   path,
	}
}

type BearerAuthenticator = func(bearer string) (api.AccessTokenJwtClaims, error)

type accessTokenKey struct{}

func contextWithAccessToken(ctx context.Context, claims api.AccessTokenJwtClaims) context.Context {
	return context.WithValue(ctx, accessTokenKey{}, claims)
}
func AccessTokenFromContext(ctx context.Context) (api.AccessTokenJwtClaims, bool) {
	token, ok := ctx.Value(accessTokenKey{}).(api.AccessTokenJwtClaims)
	return token, ok
}

func NewJwtBearerAuthenticator(jwtPublicKey string) (BearerAuthenticator, error) {
	jwtProv, err := jwt_auth.NewJwtProviderBuilder().WithPublicKey([]byte(jwtPublicKey)).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build jwt provider for jwt bearer authorizer: %w", err)
	}

	return func(bearer string) (api.AccessTokenJwtClaims, error) {
		var claims api.AccessTokenJwtClaims

		if err := jwtProv.Parse(bearer, &claims); err != nil {
			return claims, err
		}

		if claims.TokenType != "access" {
			return claims, fmt.Errorf(`invalid token type "%s", must be "access"`, claims.TokenType)
		}

		return claims, nil
	}, nil
}

type AuthMiddlewareBuilder struct {
	routes              []RequiredRoute
	bearerAuthenticator BearerAuthenticator
	logger              *zap.Logger
}

func NewBuilder() *AuthMiddlewareBuilder {
	return &AuthMiddlewareBuilder{}
}

func (b *AuthMiddlewareBuilder) Routes(routes ...RequiredRoute) *AuthMiddlewareBuilder {
	b.routes = routes
	return b
}

func (b *AuthMiddlewareBuilder) Authenticator(a BearerAuthenticator) *AuthMiddlewareBuilder {
	b.bearerAuthenticator = a
	return b
}

func (b *AuthMiddlewareBuilder) Build() (func(c *gin.Context), error) {
	if len(b.routes) == 0 {
		return nil, errors.New("auth middleware routes length must be greater than 0")
	}

	if b.bearerAuthenticator == nil {
		return nil, errors.New("auth middleware bearer authenticator must be provided")
	}

	if b.logger == nil {
		b.logger = zap.NewNop()
	}

	return func(c *gin.Context) {
		if checkAuthRequired(b.routes, c) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				b.logger.Info("authorization header value was not provided")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"errors": []gin.H{{"code": 121, "message": "Authorization header must be provided"}},
				})
				return
			}
			bearer, stripped := strings.CutPrefix(authHeader, "Bearer ")
			if !stripped {
				b.logger.Info("authorization bearer header was not provided", zap.String("value", authHeader))
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"errors": []gin.H{{"code": 122, "message": `authorization header "Bearer " prefix must be provided`}},
				})
				return
			}

			accessToken, err := b.bearerAuthenticator(bearer)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"errors": []gin.H{{"code": 123, "message": fmt.Sprintf("invalid bearer jwt access token: %v", err)}},
				})
				return
			}
			c.Request = c.Request.WithContext(contextWithAccessToken(c.Request.Context(), accessToken))
		}

		c.Next()
	}, nil
}

func checkAuthRequired(routes []RequiredRoute, c *gin.Context) bool {
	return slices.ContainsFunc(routes, func(r RequiredRoute) bool {
		return c.FullPath() == r.Path && c.Request.Method == r.Method
	})
}
