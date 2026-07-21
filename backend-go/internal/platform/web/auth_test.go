package web

import (
	"testing"
	"time"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func TestTokenRoundTrip(t *testing.T) {
	tm := NewTokenMaker(testSecret, time.Hour)
	token, err := tm.New(42, "instructor")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	id, err := tm.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if id.UserID != 42 || id.Role != "instructor" {
		t.Fatalf("got %+v, want {42 instructor}", id)
	}
}

func TestTokenRejectsWrongSecret(t *testing.T) {
	token, err := NewTokenMaker(testSecret, time.Hour).New(1, "student")
	if err != nil {
		t.Fatal(err)
	}
	other := NewTokenMaker("ffffffffffffffffffffffffffffffff", time.Hour)
	if _, err := other.Parse(token); err == nil {
		t.Fatal("expected error for token signed with a different secret")
	}
}

func TestTokenRejectsExpired(t *testing.T) {
	tm := NewTokenMaker(testSecret, -time.Minute) // already expired
	token, err := tm.New(1, "student")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tm.Parse(token); err == nil {
		t.Fatal("expected error for expired token")
	}
}
