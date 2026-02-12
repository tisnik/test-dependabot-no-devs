package jwt

import (
	"fmt"
	"strings"
	"time"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

// Authorizer verifies Bearer JWTs signed with APP_JWT_SECRET and returns principals from the token claim.
type Authorizer struct {
	secret []byte
}

func NewAuthorizer(secret string) *Authorizer {
	return &Authorizer{secret: []byte(secret)}
}

var _ auth.Authorizer = (*Authorizer)(nil)

// Claims holds application JWT claims (principals list).
type Claims struct {
	jwt.RegisteredClaims
	Principals []string `json:"principals"`
}

// Authorize implements auth.Authorizer.
func (a *Authorizer) Authorize(authorizationHeader string) ([]string, error) {
	if strings.TrimSpace(authorizationHeader) == "" {
		return nil, fmt.Errorf("missing authorization header")
	}
	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return nil, fmt.Errorf("invalid auth token syntax")
	}
	tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")
	if tokenString == "" {
		return nil, fmt.Errorf("invalid auth token syntax")
	}
	if len(a.secret) == 0 {
		return nil, fmt.Errorf("jwt not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}
	if len(claims.Principals) == 0 {
		return nil, fmt.Errorf("token has no principals")
	}
	return claims.Principals, nil
}
