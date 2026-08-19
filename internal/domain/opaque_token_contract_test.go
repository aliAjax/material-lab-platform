package domain

import (
	"errors"
	"testing"
)

func TestOpaqueTokenParserRejectsAmbiguousBoundaries(t *testing.T) {
	for _, raw := range []string{"", "missing-separator", ".secret", "session.", "session.secret.extra"} {
		if _, _, err := SplitOpaqueToken(raw); !errors.Is(err, ErrValidation) {
			t.Fatalf("SplitOpaqueToken(%q) error = %v, want validation", raw, err)
		}
	}
	sessionID, secret, err := SplitOpaqueToken("session.secret")
	if err != nil || sessionID != "session" || secret != "secret" {
		t.Fatalf("valid token = %q, %q, %v", sessionID, secret, err)
	}
}
