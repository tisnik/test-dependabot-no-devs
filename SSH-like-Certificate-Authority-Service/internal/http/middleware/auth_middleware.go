package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth"
)

type principalsContextKey struct{}

func withPrincipals(ctx context.Context, principals []string) context.Context {
	return context.WithValue(ctx, principalsContextKey{}, principals)
}

// PrincipalsFromContext returns principals set by the auth middleware, if any.
func PrincipalsFromContext(ctx context.Context) ([]string, bool) {
	principals, ok := ctx.Value(principalsContextKey{}).([]string)
	return principals, ok
}

type middlewareErrorResponse struct {
	Error string `json:"error"`
}

// AuthenticationMiddleware validates requests and stores principals in context.
type AuthenticationMiddleware struct {
	Log        *slog.Logger
	Authorizer auth.Authorizer
}

func NewMiddleware(logger *slog.Logger, authorizer auth.Authorizer) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{Log: logger, Authorizer: authorizer}
}

func (m *AuthenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m == nil || m.Authorizer == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(middlewareErrorResponse{Error: "auth middleware not configured"})
			return
		}

		principals, err := m.Authorizer.Authorize(r.Header.Get("Authorization"))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			if encErr := json.NewEncoder(w).Encode(middlewareErrorResponse{Error: err.Error()}); encErr != nil {
				if m.Log != nil {
					m.Log.Error("failed to encode auth error response", "error", encErr)
				}
			}
			return
		}

		ctx := withPrincipals(r.Context(), principals)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
