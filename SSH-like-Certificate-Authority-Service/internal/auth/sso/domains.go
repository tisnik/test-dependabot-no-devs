package sso

import (
	"errors"
	"fmt"
	"strings"

	"github.com/badoux/checkmail"
)

// IsAllowedEmailDomain reports whether the email is well-formed and its domain
// is in the allowlist (case-insensitive).
func IsAllowedEmailDomain(email string, allowed map[string]struct{}) bool {
	if email == "" || len(allowed) == 0 {
		return false
	}
	email = strings.TrimSpace(strings.ToLower(email))
	if err := checkmail.ValidateFormat(email); err != nil {
		return false
	}
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return false
	}
	domain := email[at+1:]
	if domain == "" {
		return false
	}
	_, ok := allowed[domain]
	return ok
}

var ErrEmptyAllowedDomains = errors.New("SSO_ALLOWED_DOMAINS is empty or has no valid domains")

// ParseAllowedDomainsEnv parses a comma-separated list of domains into a
// trimmed, lowercased set. It returns ErrEmptyAllowedDomains when empty.
func ParseAllowedDomainsEnv(envValue string) (map[string]struct{}, error) {
	if strings.TrimSpace(envValue) == "" {
		return nil, fmt.Errorf("%w", ErrEmptyAllowedDomains)
	}
	parts := strings.Split(envValue, ",")
	s := make(map[string]struct{})
	for _, p := range parts {
		d := strings.ToLower(strings.TrimSpace(p))
		if d != "" {
			s[d] = struct{}{}
		}
	}
	if len(s) == 0 {
		return nil, fmt.Errorf("%w", ErrEmptyAllowedDomains)
	}
	return s, nil
}
