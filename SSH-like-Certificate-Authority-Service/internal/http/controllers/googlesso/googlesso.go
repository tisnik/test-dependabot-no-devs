package googlesso

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth/jwt"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth/sso"
	gojwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

const (
	oauthStateCookieName = "oauth_state"
	oauthStateCookiePath = "/"
	oauthStateMaxAge     = 600
	appJWTExpiry         = time.Hour
)

type GoogleSSOController struct {
	Log               *slog.Logger
	Config            *oauth2.Config
	JWTSecret         []byte
	AllowedDomains    map[string]struct{}
	stateCookieSecure bool
}

type GoogleSSOConfig struct {
	ClientID       string
	ClientSecret   string
	RedirectURL    string
	JWTSecret      string
	AllowedDomains map[string]struct{}
}

func NewController(logger *slog.Logger, cfg GoogleSSOConfig) *GoogleSSOController {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	secure := strings.HasPrefix(strings.ToLower(cfg.RedirectURL), "https://")
	return &GoogleSSOController{
		Log:               logger,
		Config:            oauthCfg,
		JWTSecret:         []byte(cfg.JWTSecret),
		AllowedDomains:    cfg.AllowedDomains,
		stateCookieSecure: secure,
	}
}

// Login handles GET /auth/google/login: sets state cookie and redirects to Google.
func (c *GoogleSSOController) Login(w http.ResponseWriter, r *http.Request) {
	if c.Config.ClientID == "" || c.Config.ClientSecret == "" {
		c.Log.Warn("Google SSO not configured: missing client id or secret")
		writeJSONError(w, http.StatusServiceUnavailable, "SSO not configured")
		return
	}
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		c.Log.Error("failed to generate state", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     oauthStateCookiePath,
		MaxAge:   oauthStateMaxAge,
		HttpOnly: true,
		Secure:   c.stateCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	url := c.Config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

// Callback handles GET /auth/google/callback.
func (c *GoogleSSOController) Callback(w http.ResponseWriter, r *http.Request) {
	if len(c.JWTSecret) == 0 {
		c.Log.Warn("Google SSO callback: APP_JWT_SECRET not set")
		writeJSONError(w, http.StatusServiceUnavailable, "SSO not configured")
		return
	}

	stateCookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		c.Log.Warn("callback: missing state cookie", "error", err)
		writeJSONError(w, http.StatusBadRequest, "missing state")
		return
	}
	stateParam := r.URL.Query().Get("state")
	if stateParam == "" {
		writeJSONError(w, http.StatusBadRequest, "missing state")
		return
	}
	if stateCookie.Value != stateParam {
		c.Log.Warn("callback: state mismatch")
		writeJSONError(w, http.StatusBadRequest, "invalid state")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     oauthStateCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.stateCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		writeJSONError(w, http.StatusBadRequest, "missing code")
		return
	}

	ctx := r.Context()
	token, err := c.Config.Exchange(ctx, code)
	if err != nil {
		c.Log.Error("token exchange failed", "error", err)
		writeJSONError(w, http.StatusUnauthorized, "exchange failed")
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		c.Log.Error("id_token missing in token response")
		writeJSONError(w, http.StatusUnauthorized, "no id token")
		return
	}

	payload, err := idtoken.Validate(ctx, rawIDToken, c.Config.ClientID)
	if err != nil {
		c.Log.Error("id token validation failed", "error", err)
		writeJSONError(w, http.StatusUnauthorized, "invalid id token")
		return
	}

	email, _ := payload.Claims["email"].(string)
	if email == "" {
		c.Log.Warn("id token has no email claim")
		writeJSONError(w, http.StatusForbidden, "email not in token")
		return
	}

	if len(c.AllowedDomains) == 0 {
		c.Log.Warn("SSO_ALLOWED_DOMAINS empty; rejecting all")
		writeJSONError(w, http.StatusForbidden, "email domain not allowed")
		return
	}
	if !sso.IsAllowedEmailDomain(email, c.AllowedDomains) {
		c.Log.Info("callback: domain not allowed", "email", email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{Error: "email domain not allowed"})
		return
	}

	principals := []string{email}
	appToken, err := c.issueAppJWT(principals)
	if err != nil {
		c.Log.Error("failed to issue app JWT", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
		Email string `json:"email"`
	}{Token: appToken, Email: email})
}

// Logout handles POST /auth/logout.
func (c *GoogleSSOController) Logout(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{Message: "delete token client-side"})
}

func (c *GoogleSSOController) issueAppJWT(principals []string) (string, error) {
	now := time.Now()
	claims := &jwt.Claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(now.Add(appJWTExpiry)),
			IssuedAt:  gojwt.NewNumericDate(now),
		},
		Principals: principals,
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString(c.JWTSecret)
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: msg})
}
