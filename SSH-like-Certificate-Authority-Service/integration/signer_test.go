package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth/static"
	signerctl "github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/controllers/signer"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/middleware"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/server"
	signersvc "github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/services/signer"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/crypto/ssh"
)

const (
	validED25519Key   = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDGZSOOsYBPEs9Hns0/ggxk5E78afDlAYI/OTPbGHjdJ peto_priv@pschrimp-thinkpadx1carbongen11.tpbc.csb"
	invalidED25519Key = "ssh-ed25519 AAAAC3Nza1lZDI1NTE5AAAAIDGZSOOsYBPEs9Hns0/ggxk5E78afDlAYI/OTPbGHjdJ peto_priv@pschrimp-thinkpadx1carbongen11.tpbc.csb"
	rsaKey            = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCckU6QgWBrdqWn4/Zkdb4k9CdLrPSHlsFK8j2GIEPwXwjowO1r+2Cu93MFm0IGopc+y7pwD/K3cgFZheSO/8NW4GjZoFJfEvTTjFOPMLJjSFPHq1PLekfZJxS3jP0khPn+kv5kvj4er8FGEZpRfog3uxHmNPY1E+gyG/6lLsFIcqwY+UyIVbcRZJc7i/VY994XDdNBaRyP3AIk93YzTUUmqEve6J18tvn7GPwN2/qPqB4v4L8kK9qZuk0EQrXXnA280aO4bhNekSsFMgnWheiAAErGP010/k3kUwBF/k5oseYWAOQX7sIx0T96I8vJh+yQdgNrG+q8nLVLnLMEDvnLLoAK434hURnfazsn94wsQRdvVAZsF8j0mgfaA0xQ45Ige0zh61vub3TBPX/fGvayoThOEYVL5L85b94JtowNRqFZ/haQsQ/6rwIon2jDkGaF2QgGgMcUTS13telFh6aSGbcDQKJyQTpD6XyRJ62cYD38LpfopeiOVWoCpGRFIWU= user@host"
)

var testCASigner ssh.Signer

const caPrivateKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACCyBAP9+e8EKB2+7QORl9hYLObGzf7mFtcvGqcOzCFHGwAAALiuTVzHrk1c
xwAAAAtzc2gtZWQyNTUxOQAAACCyBAP9+e8EKB2+7QORl9hYLObGzf7mFtcvGqcOzCFHGw
AAAECgg11yFm7dJL2rZPCcOvU8MLedG28Pg+Y/gc7MYsZTNbIEA/357wQoHb7tA5GX2Fgs
5sbN/uYW1y8apw7MIUcbAAAAMXBldG9fcHJpdkBwc2NocmltcC10aGlua3BhZHgxY2FyYm
9uZ2VuMTEudHBiYy5jc2IBAgME
-----END OPENSSH PRIVATE KEY-----`

func TestMain(m *testing.M) {
	signer, err := ssh.ParsePrivateKey([]byte(caPrivateKey))
	if err != nil {
		panic("failed to parse CA private key: " + err.Error())
	}

	testCASigner = signer

	m.Run()
}

func setupTestServer(t *testing.T, allowedTokens map[string][]string) *httptest.Server {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	signingSvc := signersvc.NewSSHService(testCASigner)
	controller := signerctl.NewController(logger, signingSvc)

	authorizer := static.NewAuthorizer(allowedTokens)
	authMiddleware := middleware.NewMiddleware(logger, authorizer)
	router := server.NewMux(controller, nil, authMiddleware.Middleware)

	ts := httptest.NewServer(router)
	t.Cleanup(func() { ts.Close() })

	return ts
}

type successResponse struct {
	SignedCert string `json:"signed_cert,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func doRequestAndParseError(t *testing.T, req *http.Request, expectedStatus int) errorResponse {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var errorResp errorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if errorResp.Error == "" {
		t.Error("expected error message in response")
	}

	return errorResp
}

func TestSign_ValidRequest(t *testing.T) {
	allowedTokens := map[string][]string{
		"super-secret-token-12345": {"user1", "admin"},
	}
	ts := setupTestServer(t, allowedTokens)

	reqBody := map[string]string{
		"public_key": validED25519Key,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", ts.URL+"/sign", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer super-secret-token-12345")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var successResp successResponse
	if err := json.Unmarshal(body, &successResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if successResp.SignedCert == "" {
		t.Error("expected signed_cert in response")
	}

	// Verify SignedCert is valid SSH public key
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(successResp.SignedCert))
	if err != nil {
		t.Fatalf("failed to parse signed certificate: %v", err)
	}

	sshCert, ok := key.(*ssh.Certificate)
	if !ok {
		t.Fatal("expected SSH certificate")
	}

	// Verify principals
	expectedPrincipals := map[string]struct{}{"user1": {}, "admin": {}}
	principalMap := make(map[string]struct{})
	for _, p := range sshCert.ValidPrincipals {
		principalMap[p] = struct{}{}
	}
	if !cmp.Equal(principalMap, expectedPrincipals) {
		t.Errorf("principals mismatch: got %v, want %v", principalMap, expectedPrincipals)
	}
}

func TestSign_InvalidHTTPMethod(t *testing.T) {
	allowedTokens := map[string][]string{
		"super-secret-token-12345": {"user1", "admin"},
	}
	ts := setupTestServer(t, allowedTokens)

	req, _ := http.NewRequest("GET", ts.URL+"/sign", nil)
	req.Header.Set("Authorization", "Bearer super-secret-token-12345")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", resp.StatusCode)
	}
}

func TestSign_InvalidToken(t *testing.T) {
	allowedTokens := map[string][]string{
		"super-secret-token-12345": {"user1", "admin"},
	}
	ts := setupTestServer(t, allowedTokens)

	reqBody := map[string]string{
		"public_key": validED25519Key,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", ts.URL+"/sign", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer invalid_token")
	req.Header.Set("Content-Type", "application/json")

	doRequestAndParseError(t, req, http.StatusUnauthorized)
}

func TestSign_MissingAuthorizationHeader(t *testing.T) {
	allowedTokens := map[string][]string{
		"super-secret-token-12345": {"user1", "admin"},
	}
	ts := setupTestServer(t, allowedTokens)

	reqBody := map[string]string{
		"public_key": validED25519Key,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", ts.URL+"/sign", bytes.NewBuffer(jsonBody))
	// Intentionally not setting Authorization header
	req.Header.Set("Content-Type", "application/json")

	doRequestAndParseError(t, req, http.StatusUnauthorized)
}

func TestSign_MissingBearerPrefix(t *testing.T) {
	allowedTokens := map[string][]string{
		"super-secret-token-12345": {"user1", "admin"},
	}
	ts := setupTestServer(t, allowedTokens)

	reqBody := map[string]string{
		"public_key": validED25519Key,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", ts.URL+"/sign", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "super-secret-token-12345") // Missing "Bearer " prefix
	req.Header.Set("Content-Type", "application/json")

	doRequestAndParseError(t, req, http.StatusUnauthorized)
}

func TestSign_InvalidJSONBody(t *testing.T) {
	allowedTokens := map[string][]string{
		"super-secret-token-12345": {"user1", "admin"},
	}
	ts := setupTestServer(t, allowedTokens)

	reqBody := map[string]string{
		"INVALID": validED25519Key,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", ts.URL+"/sign", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer super-secret-token-12345")
	req.Header.Set("Content-Type", "application/json")

	doRequestAndParseError(t, req, http.StatusBadRequest)
}

func TestSign_InvalidPublicKeyFormat(t *testing.T) {
	allowedTokens := map[string][]string{
		"super-secret-token-12345": {"user1", "admin"},
	}
	ts := setupTestServer(t, allowedTokens)

	reqBody := map[string]string{
		"public_key": invalidED25519Key, // Key with typo: Nza1 instead of NzaC
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", ts.URL+"/sign", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer super-secret-token-12345")
	req.Header.Set("Content-Type", "application/json")

	doRequestAndParseError(t, req, http.StatusBadRequest)
}

func TestSign_WrongKeyType(t *testing.T) {
	allowedTokens := map[string][]string{
		"super-secret-token-12345": {"user1", "admin"},
	}
	ts := setupTestServer(t, allowedTokens)

	reqBody := map[string]string{
		"public_key": rsaKey, // RSA key instead of ED25519
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", ts.URL+"/sign", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer super-secret-token-12345")
	req.Header.Set("Content-Type", "application/json")

	doRequestAndParseError(t, req, http.StatusBadRequest)
}
