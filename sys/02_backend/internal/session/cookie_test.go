package session_test

import (
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/session"
)

func TestSignAndVerify_RoundTrips(t *testing.T) {
	signed := session.Sign("session-id-123", "secret")
	id, ok := session.Verify(signed, "secret")
	if !ok {
		t.Fatal("Verify returned ok=false for a validly signed value")
	}
	if id != "session-id-123" {
		t.Errorf("id = %q, want session-id-123", id)
	}
}

func TestVerify_RejectsWrongSecret(t *testing.T) {
	signed := session.Sign("session-id-123", "secret")
	_, ok := session.Verify(signed, "wrong-secret")
	if ok {
		t.Error("Verify should reject a cookie signed with a different secret")
	}
}

func TestVerify_RejectsTamperedID(t *testing.T) {
	signed := session.Sign("session-id-123", "secret")
	// Swap the id portion but keep the original signature.
	tampered := "attacker-id" + signed[len("session-id-123"):]
	_, ok := session.Verify(tampered, "secret")
	if ok {
		t.Error("Verify should reject a cookie whose id was tampered with")
	}
}

func TestVerify_RejectsMalformedValue(t *testing.T) {
	_, ok := session.Verify("no-dot-separator", "secret")
	if ok {
		t.Error("Verify should reject a value with no signature separator")
	}
}

func TestRandomToken_ProducesDistinctValues(t *testing.T) {
	a, err := session.RandomToken()
	if err != nil {
		t.Fatalf("RandomToken: %v", err)
	}
	b, err := session.RandomToken()
	if err != nil {
		t.Fatalf("RandomToken: %v", err)
	}
	if a == b {
		t.Error("two calls to RandomToken produced the same value")
	}
}
