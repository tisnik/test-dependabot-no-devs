package static

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func sliceToSet(ss []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		set[s] = struct{}{}
	}
	return set
}

func TestAuthorizer_Authorize(t *testing.T) {
	allowedTokens := map[string][]string{
		"valid-token-1": {"user1", "admin"},
		"valid-token-2": {"user2"},
		"empty-token":   {},
	}

	authorizer := NewAuthorizer(allowedTokens)

	tests := []struct {
		name           string
		token          string
		expectedResult []string
		expectErr      bool
	}{
		{
			name:           "valid token with Bearer prefix",
			token:          "Bearer valid-token-1",
			expectedResult: []string{"admin", "user1"},
			expectErr:      false,
		},
		{
			name:           "valid token with different principals",
			token:          "Bearer valid-token-2",
			expectedResult: []string{"user2"},
			expectErr:      false,
		},
		{
			name:           "valid token with empty principals",
			token:          "Bearer empty-token",
			expectedResult: []string{},
			expectErr:      false,
		},
		{
			name:           "token without Bearer prefix",
			token:          "valid-token-1",
			expectedResult: nil,
			expectErr:      true,
		},
		{
			name:           "invalid token with Bearer prefix",
			token:          "Bearer invalid-token",
			expectedResult: nil,
			expectErr:      true,
		},
		{
			name:           "empty token",
			token:          "",
			expectedResult: nil,
			expectErr:      true,
		},
		{
			name:           "just Bearer prefix",
			token:          "Bearer ",
			expectedResult: nil,
			expectErr:      true,
		},
		{
			name:           "Bearer prefix with extra spaces",
			token:          "Bearer  valid-token-1",
			expectedResult: nil,
			expectErr:      true,
		},
		{
			name:           "case sensitive Bearer",
			token:          "bearer valid-token-1",
			expectedResult: nil,
			expectErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := authorizer.Authorize(tt.token)
			if err != nil && !tt.expectErr {
				t.Errorf("expected nil error, got %v", err)
				return
			}
			if err == nil && tt.expectErr {
				t.Errorf("expected error, got %v", err)
				return
			}
			if !cmp.Equal(sliceToSet(result), sliceToSet(tt.expectedResult)) {
				t.Errorf("Authorize() result = %v, expectedResult = %v", result, tt.expectedResult)
			}
		})
	}
}
