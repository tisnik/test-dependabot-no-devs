package server

import (
	"net/http"
	"time"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/controllers/googlesso"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/controllers/signer"
	"github.com/go-chi/chi/v5"
)

const (
	defaultServerReadTimeout       = 1 * time.Second
	defaultServerReadHeaderTimeout = 2 * time.Second
	defaultServerWriteTimeout      = 1 * time.Second
	defaultServerIdleTimeout       = 30 * time.Second
)

func NewServer(handler http.Handler, address string) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadTimeout:       defaultServerReadTimeout,
		ReadHeaderTimeout: defaultServerReadHeaderTimeout,
		WriteTimeout:      defaultServerWriteTimeout,
		IdleTimeout:       defaultServerIdleTimeout,
	}
}

func NewMux(sc *signer.SignerController, ssoc *googlesso.GoogleSSOController, middlewares ...func(http.Handler) http.Handler) chi.Router {
	root := chi.NewRouter()

	if ssoc != nil {
		root.Get("/auth/google/login", ssoc.Login)
		root.Get("/auth/google/callback", ssoc.Callback)
		root.Post("/auth/logout", ssoc.Logout)
	}

	protected := chi.NewRouter()
	for _, mw := range middlewares {
		protected.Use(mw)
	}
	if sc != nil {
		protected.Post("/sign", sc.Sign)
	}
	root.Mount("/", protected)

	return root
}
