package sso

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func set(domains ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(domains))
	for _, d := range domains {
		m[d] = struct{}{}
	}
	return m
}

func TestIsAllowedEmailDomain(t *testing.T) {
	allowedDomains := set("gmail.com", "redhat.com", "abc.com")

	tests := []struct {
		name           string
		email          string
		allowedDomains map[string]struct{}
		wantAllowed    bool
	}{
		{
			name:           "allowed gmail",
			email:          "user@gmail.com",
			allowedDomains: allowedDomains,
			wantAllowed:    true,
		},
		{
			name:           "allowed gmail uppercase",
			email:          "User@Gmail.COM",
			allowedDomains: allowedDomains,
			wantAllowed:    true,
		},
		{
			name:           "allowed redhat",
			email:          "foo@redhat.com",
			allowedDomains: allowedDomains,
			wantAllowed:    true,
		},
		{
			name:           "allowed redhat uppercase",
			email:          "Foo@RedHat.Com",
			allowedDomains: allowedDomains,
			wantAllowed:    true,
		},
		{
			name:           "rejected yahoo",
			email:          "user@yahoo.com",
			allowedDomains: allowedDomains,
			wantAllowed:    false,
		},
		{
			name:           "rejected example",
			email:          "user@example.com",
			allowedDomains: allowedDomains,
			wantAllowed:    false,
		},
		{
			name:           "rejected company",
			email:          "user@company.org",
			allowedDomains: allowedDomains,
			wantAllowed:    false,
		},
		{
			name:           "rejected gmail.com.evil.com",
			email:          "user@gmail.com.evil.com",
			allowedDomains: allowedDomains,
			wantAllowed:    false,
		},
		{
			name:           "rejected subdomain gmail",
			email:          "user@mail.gmail.com",
			allowedDomains: allowedDomains,
			wantAllowed:    false,
		},
		{
			name:           "empty email",
			email:          "",
			allowedDomains: allowedDomains,
			wantAllowed:    false,
		},
		{
			name:           "no at",
			email:          "usergmail.com",
			allowedDomains: allowedDomains,
			wantAllowed:    false,
		},
		{
			name:           "empty allowed domains",
			email:          "user@gmail.com",
			allowedDomains: nil,
			wantAllowed:    false,
		},
		{
			name:           "empty allowed domains set",
			email:          "user@gmail.com",
			allowedDomains: map[string]struct{}{},
			wantAllowed:    false,
		},
		{
			name:           "at only",
			email:          "@",
			allowedDomains: allowedDomains,
			wantAllowed:    false,
		},
		{
			name:           "domain only",
			email:          "@gmail.com",
			allowedDomains: allowedDomains,
			wantAllowed:    false,
		},
		{
			name:           "multiple ats",
			email:          "user@gmail.com@redhat.com",
			allowedDomains: allowedDomains,
			wantAllowed:    false,
		},
		{
			name:           "case insensitive",
			email:          "user@Abc.com",
			allowedDomains: allowedDomains,
			wantAllowed:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAllowedEmailDomain(tt.email, tt.allowedDomains)
			if got != tt.wantAllowed {
				t.Errorf("IsAllowedEmailDomain(%q, %v) = %v, want %v", tt.email, tt.allowedDomains, got, tt.wantAllowed)
			}
		})
	}
}

func TestParseAllowedDomains(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     map[string]struct{}
		wantErr  bool
	}{
		{
			name:     "normal",
			envValue: "gmail.com,redhat.com",
			want:     set("gmail.com", "redhat.com"),
			wantErr:  false,
		},
		{
			name:     "empty",
			envValue: "",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "whitespace only",
			envValue: "  ,  ",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "trim and lowercase",
			envValue: "  Gmail.COM , RedHat.COM  ",
			want:     set("gmail.com", "redhat.com"),
			wantErr:  false,
		},
		{
			name:     "dedupe",
			envValue: "gmail.com,gmail.com,redhat.com",
			want:     set("gmail.com", "redhat.com"),
			wantErr:  false,
		},
		{
			name:     "dedupe single",
			envValue: "aAa.sk",
			want:     set("aaa.sk"),
			wantErr:  false,
		},
		{
			name:     "dedupe case",
			envValue: "Aaa.sk,aAa.sk",
			want:     set("aaa.sk"),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAllowedDomainsEnv(tt.envValue)
			if tt.wantErr && err == nil {
				t.Errorf("ParseAllowedDomainsEnv(%q) expected error", tt.envValue)
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ParseAllowedDomainsEnv(%q) unexpected error: %v", tt.envValue, err)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("ParseAllowedDomainsEnv(%q) = %v, want %v", tt.envValue, got, tt.want)
			}
		})
	}
}
