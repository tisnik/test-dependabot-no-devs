package main

import (
	"log/slog"
	"os"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth/jwt"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth/multi"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth/sso"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth/static"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/controllers/googlesso"
	signerctl "github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/controllers/signer"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/middleware"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/server"
	signersvc "github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/services/signer"
	"golang.org/x/crypto/ssh"
)

const (
	defaultListenAddr   = ":8443"
	baseURL             = "https://localhost" + defaultListenAddr
	appSecretsDir       = "/run/ca-service"
	privateHTTPSKeyFile = appSecretsDir + "/https/ca-service-local.key.pem"
	publicHTTPSCertFile = appSecretsDir + "/https/ca-service-local.cert.pem"
	privateKeyFile      = appSecretsDir + "/ssh/ca_key"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	allowedTokens, staticErr := static.ParseStaticTokenPrincipals(os.Getenv("CA_ACCESS_TOKEN"))
	if staticErr != nil {
		logger.Error("static token parser failed", "error", staticErr, "env", "CA_ACCESS_TOKEN empty or invalid format")
	} else if len(allowedTokens) == 0 {
		logger.Warn("No static tokens parsed from CA_ACCESS_TOKEN")
	} else {
		logger.Info("static tokens parsed from CA_ACCESS_TOKEN", "count", len(allowedTokens))
	}

	ssoDomains, ssoErr := sso.ParseAllowedDomainsEnv(os.Getenv("SSO_ALLOWED_DOMAINS"))
	if ssoErr != nil {
		logger.Error("SSO allowed domains parser failed", "error", ssoErr, "env", "SSO_ALLOWED_DOMAINS empty or no valid domains")
	} else if len(ssoDomains) == 0 {
		logger.Warn("No allowed domains parsed from SSO_ALLOWED_DOMAINS")
	} else {
		logger.Info("SSO allowed domains parsed from SSO_ALLOWED_DOMAINS", "count", len(ssoDomains))
	}

	if staticErr != nil && ssoErr != nil {
		logger.Error("both parsers for static tokens and SSO domains failed")
		os.Exit(1)
	}

	caKeyBytes, err := os.ReadFile(privateKeyFile)
	if err != nil {
		logger.Error("failed to read CA private key", "error", err)
		os.Exit(1)
	}

	caSigner, err := ssh.ParsePrivateKey(caKeyBytes)
	if err != nil {
		logger.Error("failed to parse CA private key", "error", err)
		os.Exit(1)
	}

	signingSvc := signersvc.NewSSHService(caSigner)
	controller := signerctl.NewController(logger.With("component", "api"), signingSvc)

	staticAuth := static.NewAuthorizer(allowedTokens)
	jwtAuth := jwt.NewAuthorizer(os.Getenv("APP_JWT_SECRET"))
	authorizer := multi.NewAuthorizer(jwtAuth, staticAuth)
	authMiddleware := middleware.NewMiddleware(logger.With("component", "auth"), authorizer)

	googleSSO := googlesso.NewController(logger.With("component", "sso"), googlesso.GoogleSSOConfig{
		ClientID:       os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:    baseURL + "/auth/google/callback",
		JWTSecret:      os.Getenv("APP_JWT_SECRET"),
		AllowedDomains: ssoDomains, // nil when SSO parser failed; callback will reject
	})

	root := server.NewMux(controller, googleSSO, authMiddleware.Middleware)

	controller.Log.Info("service starting", "base URL", baseURL)

	s := server.NewServer(root, defaultListenAddr)

	err = s.ListenAndServeTLS(publicHTTPSCertFile, privateHTTPSKeyFile)
	if err != nil {
		logger.Error("Error with listening and serving port", "error", err)
		os.Exit(1)
	}
}
