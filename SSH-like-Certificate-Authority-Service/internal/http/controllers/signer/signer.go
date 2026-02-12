package signer

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/http/middleware"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/services/signer"
	"golang.org/x/crypto/ssh"
)

type SignerController struct {
	Log    *slog.Logger
	signer signer.Signer
}

type requestBody struct {
	PublicKey string `json:"public_key"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type successResponse struct {
	SignedCert string `json:"signed_cert,omitempty"`
}

func NewController(logger *slog.Logger, signer signer.Signer) *SignerController {
	return &SignerController{Log: logger, signer: signer}
}

// Sign handles POST /sign: returns a signed SSH certificate.
func (c *SignerController) Sign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	principals, ok := middleware.PrincipalsFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		if encErr := json.NewEncoder(w).Encode(errorResponse{Error: "unauthorized"}); encErr != nil {
			c.Log.Error("failed to encode error response", "error", encErr)
		}
		return
	}

	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(errorResponse{Error: "failed to decode request body"}); encErr != nil {
			c.Log.Error("failed to encode error response", "error", encErr)
		}
		return
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(reqBody.PublicKey))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(errorResponse{Error: "failed to parse public key"}); encErr != nil {
			c.Log.Error("failed to encode error response", "error", encErr)
		}
		return
	}

	if pubKey.Type() != ssh.KeyAlgoED25519 {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(errorResponse{Error: "only ed25519 keys are supported"}); encErr != nil {
			c.Log.Error("failed to encode error response", "error", encErr)
		}
		return
	}

	signedCert, err := c.signer.SignUserKey(pubKey, principals)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(errorResponse{Error: "failed to sign certificate"}); encErr != nil {
			c.Log.Error("failed to encode error response", "error", encErr)
		}
		return
	}

	issuedAt := time.Unix(int64(signedCert.ValidAfter), 0)
	expiresAt := time.Unix(int64(signedCert.ValidBefore), 0)
	c.Log.Info("certificate issued",
		"issued_at", issuedAt.Format(time.RFC3339),
		"principals", principals,
		"expires_at", expiresAt.Format(time.RFC3339),
		"serial", signedCert.Serial,
	)

	w.WriteHeader(http.StatusOK)
	certBytes := ssh.MarshalAuthorizedKey(signedCert)
	if err := json.NewEncoder(w).Encode(successResponse{SignedCert: string(certBytes)}); err != nil {
		c.Log.Error("failed to encode success response", "error", err)
	}
}
