package multi

import (
	"fmt"
	"strings"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth"
)

// Authorizer tries a list of authorizers in order; the first success is returned.
type Authorizer struct {
	authorizers []auth.Authorizer
}

func NewAuthorizer(authorizers ...auth.Authorizer) *Authorizer {
	return &Authorizer{authorizers: authorizers}
}

// Authorize implements auth.Authorizer.
func (m *Authorizer) Authorize(authorizationHeader string) ([]string, error) {
	var errs []error
	for _, a := range m.authorizers {
		if a == nil {
			continue
		}
		principals, err := a.Authorize(authorizationHeader)
		if err == nil {
			return principals, nil
		}
		errs = append(errs, fmt.Errorf("%T: %w", a, err))
	}
	return nil, unauthorizedErrors(errs)
}

func unauthorizedErrors(errs []error) error {
	if len(errs) == 0 {
		return fmt.Errorf("unauthorized")
	}
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, e.Error())
	}
	return fmt.Errorf("unauthorized: %s", strings.Join(msgs, " | "))
}
