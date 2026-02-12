package static

import (
	"fmt"
	"strings"
)

// ParseStaticTokenPrincipals parses CA_ACCESS_TOKEN into a token -> principals map.
func ParseStaticTokenPrincipals(envTokens string) (map[string][]string, error) {
	if strings.TrimSpace(envTokens) == "" {
		return map[string][]string{}, fmt.Errorf("CA_ACCESS_TOKEN is empty")
	}
	m := make(map[string][]string)

	entries := strings.Split(envTokens, ";")
	for _, entry := range entries {
		parts := strings.Split(entry, ":")
		if len(parts) != 2 {
			return map[string][]string{}, fmt.Errorf("invalid token principals format")
		}

		token := parts[0]
		if token == "" {
			return map[string][]string{}, fmt.Errorf("invalid token principals format")
		}
		principals := strings.Split(parts[1], ",")
		m[token] = principals
	}

	return m, nil
}
