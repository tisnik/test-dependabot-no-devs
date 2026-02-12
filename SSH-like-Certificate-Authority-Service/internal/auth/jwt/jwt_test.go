package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
)

func TestAuthorizer_ValidToken(t *testing.T) {
	secret := "test-secret"
	authorizer := NewAuthorizer(secret)
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Principals: []string{"user@gmail.com"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	principals, err := authorizer.Authorize("Bearer " + tokenString)
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	if !cmp.Equal(principals, []string{"user@gmail.com"}) {
		t.Errorf("principals = %v, want [user@gmail.com]", principals)
	}
}

func TestAuthorizer_InvalidOrExpiredToken(t *testing.T) {
	authorizer := NewAuthorizer("secret")
	tests := []string{
		"Bearer invalid.jwt.token",
		"Bearer ",
		"not-bearer",
	}
	for _, h := range tests {
		_, err := authorizer.Authorize(h)
		if err == nil {
			t.Errorf("Authorize(%q) expected error", h)
		}
	}
}
