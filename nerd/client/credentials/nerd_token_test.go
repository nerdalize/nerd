package credentials

import (
	"strings"
	"testing"
)

const key = `
-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEAkYbLnam4wo+heLlTZEeh1ZWsfruz9nk
kyvc4LwKZ8pez5KYY76H1ox+AfUlWOEq+bExypcFfEIrJkf/JXa7jpzkOWBDF9Sa
OWbQHMK+vvUXieCJvCc9Vj084ABwLBgX
-----END PUBLIC KEY-----
`

func TestParseECDSAPublicKeyFromPemBytes(t *testing.T) {
	_, err := ParseECDSAPublicKeyFromPemBytes([]byte(key))
	if err != nil {
		t.Errorf("Failed to parse valid public key. Error message: %v", err)
	}
}

func TestDecodeToken(t *testing.T) {
	testCases := map[string]struct {
		token     string
		success   bool
		errorMsg  string
		expiresAt int64
		audience  string
	}{
		"expired valid": {
			token:     "eyJhbGciOiJFUzM4NCIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhdXRoLm5lcmRhbGl6ZS5jb20iLCJleHAiOjE0ODgxODk1MzcsImF1ZCI6Im5jZS5uZXJkYWxpemUuY29tIiwibmJmIjoxNDg4MTg5MTc3LCJhY2Nlc3MiOlt7InNlcnZpY2UiOiJuY2UubmVyZGFsaXplLmNvbSIsInJlc291cmNlX3R5cGUiOiJjbHVzdGVyIiwicmVzb3VyY2VfaWRlbnRpZmllciI6ImNsdXN0ZXIxLm5lcmQubmV0L2UxNWxqMTJhYWZzZCIsImFjdGlvbnMiOiJDUlVEIn0seyJzZXJ2aWNlIjoibmNlLm5lcmRhbGl6ZS5jb20iLCJyZXNvdXJjZV90eXBlIjoib2JqZWN0X3N0b3JlIiwicmVzb3VyY2VfaWRlbnRpZmllciI6Im5lcmRzLnMzLmFtYXphbmF3cy5jb20vMjQ1bGtqMjM0NSIsImFjdGlvbnMiOiJSIn1dLCJzdWIiOiI0IiwiaWF0IjoxNDg4MTg5MjM3fQ.pk7yAGW8L80uEKAFUctupj4PO8UHIGmpikEi-ERkwZao73dEx5GlAnVmNTnXOO-xxjT8BomQtqL6Od15d7K6c4fY5YU8s64di4HA1SJuqIK0u0Mk8N6oVS216Y3FJkkD",
			success:   true,
			errorMsg:  "",
			expiresAt: 1488189537,
			audience:  "nce.nerdalize.com",
		},
		"json parse error": {
			token:     "jwt.jwt.jwt",
			success:   false,
			errorMsg:  "failed to parse nerd token",
			expiresAt: 111,
			audience:  "nlz.com",
		},
	}

	for name, tc := range testCases {
		claims, err := DecodeTokenWithPEM(tc.token, key)
		if tc.success {
			if err != nil {
				t.Errorf("%v: expected success but got error '%v'", name, err)
				continue
			}
			if claims.Audience != tc.audience {
				t.Errorf("%v: expected audience to be '%v' but was '%v'", name, tc.audience, claims.Audience)
			}
			if claims.ExpiresAt != tc.expiresAt {
				t.Errorf("%v: expected expiresAt to be '%v' but was '%v'", name, tc.expiresAt, claims.ExpiresAt)
			}
		} else {
			if err == nil {
				t.Errorf("%v: expected failure but got success", name)
				continue
			}
			if !strings.Contains(err.Error(), tc.errorMsg) {
				t.Errorf("%v: expected error message to contain '%v' but error message was '%v'", name, tc.errorMsg, err.Error())
			}
		}
	}
}
