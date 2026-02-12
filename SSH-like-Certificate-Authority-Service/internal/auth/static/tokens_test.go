package static

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseStaticTokenPrincipals(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult map[string][]string
		expectErr      bool
	}{
		{
			name:           "valid single token with single principal",
			input:          "token1:user1",
			expectedResult: map[string][]string{"token1": {"user1"}},
			expectErr:      false,
		},
		{
			name:           "valid single token with multiple principals",
			input:          "token1:user1,admin,root",
			expectedResult: map[string][]string{"token1": {"user1", "admin", "root"}},
			expectErr:      false,
		},
		{
			name:  "valid multiple tokens",
			input: "token1:user1,admin;token2:user2",
			expectedResult: map[string][]string{
				"token1": {"user1", "admin"},
				"token2": {"user2"},
			},
			expectErr: false,
		},
		{
			name:           "empty string",
			input:          "",
			expectedResult: nil,
			expectErr:      true,
		},
		{
			name:           "invalid format - missing colon",
			input:          "token1user1",
			expectedResult: nil,
			expectErr:      true,
		},
		{
			name:           "invalid format - multiple colons",
			input:          "token1:user1:extra",
			expectedResult: nil,
			expectErr:      true,
		},
		{
			name:           "empty token",
			input:          ":user1",
			expectedResult: nil,
			expectErr:      true,
		},
		{
			name:           "empty principals",
			input:          "token1:",
			expectedResult: map[string][]string{"token1": {""}},
			expectErr:      false,
		},
		{
			name:           "token with empty principal in list",
			input:          "token1:user1,,user2",
			expectedResult: map[string][]string{"token1": {"user1", "", "user2"}},
			expectErr:      false,
		},
		{
			name:  "multiple tokens with various formats",
			input: "token1:user1;token2:user2,admin;token3:root",
			expectedResult: map[string][]string{
				"token1": {"user1"},
				"token2": {"user2", "admin"},
				"token3": {"root"},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseStaticTokenPrincipals(tt.input)
			if err != nil {
				if !tt.expectErr {
					t.Errorf("expected nil error, got %v", err)
				}
				return
			}

			if tt.expectErr {
				t.Errorf("expected error, got %v", err)
				return
			}

			if !cmp.Equal(result, tt.expectedResult) {
				t.Errorf("getPrincipals() result = %v, expectedResult = %v", result, tt.expectedResult)
				return
			}
		})
	}
}
