package signer

import (
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	DefaultCertValidity = 30 * time.Minute
)

// Signer issues SSH certificates for user public keys.
type Signer interface {
	SignUserKey(userPubKey ssh.PublicKey, principals []string) (*ssh.Certificate, error)
}

type SSHSigner struct {
	CASigner ssh.Signer
	Validity time.Duration
	Now      func() time.Time
	Rand     io.Reader
}

func NewSSHService(caSigner ssh.Signer) *SSHSigner {
	return &SSHSigner{
		CASigner: caSigner,
		Validity: DefaultCertValidity,
		Now:      time.Now,
		Rand:     rand.Reader,
	}
}

func (s *SSHSigner) SignUserKey(userPubKey ssh.PublicKey, principals []string) (*ssh.Certificate, error) {
	if s == nil || s.CASigner == nil {
		return nil, fmt.Errorf("CA signer not configured")
	}

	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	validity := s.Validity
	if validity == 0 {
		validity = DefaultCertValidity
	}
	randReader := s.Rand
	if randReader == nil {
		randReader = rand.Reader
	}

	cert := &ssh.Certificate{
		Key:             userPubKey,
		CertType:        ssh.UserCert,
		Serial:          uint64(now.UnixNano()),
		ValidPrincipals: principals,
		ValidAfter:      uint64(now.Unix()),
		ValidBefore:     uint64(now.Add(validity).Unix()),
	}

	if err := cert.SignCert(randReader, s.CASigner); err != nil {
		return nil, err
	}

	return cert, nil
}
