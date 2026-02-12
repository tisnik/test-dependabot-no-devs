package auth

// Authorizer validates an Authorization header and returns principals.
type Authorizer interface {
	Authorize(authorizationHeader string) ([]string, error)
}
