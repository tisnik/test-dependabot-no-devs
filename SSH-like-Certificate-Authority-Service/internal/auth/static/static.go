package static

import (
	"fmt"
	"strings"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth"
)

// Authorizer validates Bearer tokens against a static map of token -> principals.
type Authorizer struct {
	allowedTokens map[string][]string
}

func NewAuthorizer(allowedTokens map[string][]string) *Authorizer {
	return &Authorizer{allowedTokens: allowedTokens}
}

var _ auth.Authorizer = (*Authorizer)(nil)

// Authorize implements auth.Authorizer.
func (a *Authorizer) Authorize(authorizationHeader string) ([]string, error) {
	if strings.HasPrefix(authorizationHeader, "Bearer ") {
		authorizationHeader = strings.TrimPrefix(authorizationHeader, "Bearer ")
	} else {
		return nil, fmt.Errorf("invalid auth token syntax")
	}

	principals, exists := a.allowedTokens[authorizationHeader]
	if !exists {
		return nil, fmt.Errorf("access token not valid or has no principals")
	}
	return principals, nil
}
