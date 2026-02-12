package multi

import (
	"testing"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth/jwt"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/auth/static"
	"github.com/google/go-cmp/cmp"
)

func sliceToSet(ss []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		set[s] = struct{}{}
	}
	return set
}

func TestAuthorizer_JWTFailStaticSucceeds(t *testing.T) {
	jwtAuth := jwt.NewAuthorizer("")
	staticAuth := static.NewAuthorizer(map[string][]string{
		"static-token": {"alice", "bob"},
	})
	multi := NewAuthorizer(jwtAuth, staticAuth)

	principals, err := multi.Authorize("Bearer static-token")
	if err != nil {
		t.Fatalf("expected success via static fallback: %v", err)
	}
	if !cmp.Equal(sliceToSet(principals), sliceToSet([]string{"alice", "bob"})) {
		t.Errorf("principals = %v, want [alice bob]", principals)
	}
}

func TestAuthorizer_BothFail(t *testing.T) {
	jwtAuth := jwt.NewAuthorizer("")
	staticAuth := static.NewAuthorizer(map[string][]string{"other": {"x"}})
	multi := NewAuthorizer(jwtAuth, staticAuth)

	_, err := multi.Authorize("Bearer unknown-token")
	if err == nil {
		t.Error("expected error when both authorizers reject")
	}
}
