package integration

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth/sso"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/controllers/googlesso"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/server"
)

const (
	oauthStateCookieName = "oauth_state"
	redirectURL          = "https://localhost:8443/auth/google/callback"
)

func getSSOConfigFromEnv(t *testing.T) (googlesso.GoogleSSOConfig, bool) {
	t.Helper()
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	jwtSecret := os.Getenv("APP_JWT_SECRET")
	if clientID == "" || clientSecret == "" || jwtSecret == "" {
		return googlesso.GoogleSSOConfig{}, false
	}
	allowedDomains, err := sso.ParseAllowedDomainsEnv(os.Getenv("SSO_ALLOWED_DOMAINS"))
	if err != nil {
		return googlesso.GoogleSSOConfig{}, false
	}
	return googlesso.GoogleSSOConfig{
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		RedirectURL:    redirectURL,
		JWTSecret:      jwtSecret,
		AllowedDomains: allowedDomains,
	}, true
}

func setupSSOTestServer(t *testing.T, cfg googlesso.GoogleSSOConfig) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	googleSSO := googlesso.NewController(logger, cfg)
	router := server.NewMux(nil, googleSSO)
	ts := httptest.NewServer(router)
	t.Cleanup(func() { ts.Close() })
	return ts
}

func TestSSO_Login_RedirectsToGoogle(t *testing.T) {
	cfg, ok := getSSOConfigFromEnv(t)
	if !ok {
		t.Skip("SSO env not configured")
	}
	ts := setupSSOTestServer(t, cfg)
	client := ts.Client()
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		// Don't follow redirects; we want to assert on the initial 302 + Location header.
		return http.ErrUseLastResponse
	}
	resp, err := client.Get(ts.URL + "/auth/google/login")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusFound {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 302, got %d; body=%q", resp.StatusCode, string(b))
	}
	loc := resp.Header.Get("Location")
	if loc == "" {
		t.Fatal("missing Location header")
	}
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("invalid Location URL: %v", err)
	}
	if !strings.Contains(u.Host, "google.com") {
		t.Errorf("expected redirect to Google, got host %q", u.Host)
	}
	clientID := u.Query().Get("client_id")
	wantClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID != wantClientID {
		t.Errorf("client_id: got %q, want %q", clientID, wantClientID)
	}
}

func TestSSO_Callback_MissingStateCookie_Returns400(t *testing.T) {
	cfg, ok := getSSOConfigFromEnv(t)
	if !ok {
		t.Skip("SSO env not configured")
	}
	ts := setupSSOTestServer(t, cfg)
	resp, err := http.Get(ts.URL + "/auth/google/callback")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if body.Error != "missing state" {
		t.Errorf("error: got %q, want %q", body.Error, "missing state")
	}
}

func TestSSO_Callback_StateMismatch_Returns400(t *testing.T) {
	cfg, ok := getSSOConfigFromEnv(t)
	if !ok {
		t.Skip("SSO env not configured")
	}
	ts := setupSSOTestServer(t, cfg)
	state1 := base64.URLEncoding.EncodeToString([]byte("state_value_one_32_bytes_long!!!!"))
	state2 := base64.URLEncoding.EncodeToString([]byte("state_value_two_32_bytes_long!!!!"))
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/auth/google/callback?state="+url.QueryEscape(state2), nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: state1})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if body.Error != "invalid state" {
		t.Errorf("error: got %q, want %q", body.Error, "invalid state")
	}
}

func TestSSO_Callback_MissingCode_Returns400(t *testing.T) {
	cfg, ok := getSSOConfigFromEnv(t)
	if !ok {
		t.Skip("SSO env not configured")
	}
	ts := setupSSOTestServer(t, cfg)
	state := base64.URLEncoding.EncodeToString([]byte("valid_state_32_bytes_long!!!!!!"))
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/auth/google/callback?state="+url.QueryEscape(state), nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: state})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if body.Error != "missing code" {
		t.Errorf("error: got %q, want %q", body.Error, "missing code")
	}
}

func TestSSO_Callback_InvalidCode_Returns401(t *testing.T) {
	cfg, ok := getSSOConfigFromEnv(t)
	if !ok {
		t.Skip("SSO env not configured")
	}
	ts := setupSSOTestServer(t, cfg)
	state := base64.URLEncoding.EncodeToString([]byte("valid_state_32_bytes_long!!!!!!"))
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/auth/google/callback?state="+url.QueryEscape(state)+"&code=invalid_or_expired", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: state})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if !strings.Contains(body.Error, "exchange") {
		t.Errorf("error should mention exchange: got %q", body.Error)
	}
}

func TestSSO_Logout_Returns200(t *testing.T) {
	// Logout does not depend on Google config
	cfg := googlesso.GoogleSSOConfig{
		AllowedDomains: map[string]struct{}{"example.com": {}},
	}
	ts := setupSSOTestServer(t, cfg)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/auth/logout", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if body.Message != "delete token client-side" {
		t.Errorf("message: got %q, want %q", body.Message, "delete token client-side")
	}
}
