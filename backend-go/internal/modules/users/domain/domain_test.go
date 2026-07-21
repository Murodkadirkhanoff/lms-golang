package domain

import (
	"testing"

	"github.com/chashma/lms/internal/platform/web"
)

func TestValidatePassword(t *testing.T) {
	short := web.NewValidator()
	ValidatePassword(short, "abc")
	if short.Valid() {
		t.Error("short password should be invalid")
	}

	ok := web.NewValidator()
	ValidatePassword(ok, "longenough")
	if !ok.Valid() {
		t.Errorf("expected valid, got %v", ok.Errors)
	}
}

func TestValidateEmailAndName(t *testing.T) {
	v := web.NewValidator()
	ValidateEmail(v, "not-an-email")
	ValidateName(v, "")
	if _, ok := v.Errors["email"]; !ok {
		t.Error("expected email error")
	}
	if _, ok := v.Errors["name"]; !ok {
		t.Error("expected name error")
	}
}

func TestAvatarColorDeterministic(t *testing.T) {
	if AvatarColor(5) != AvatarColor(5) {
		t.Fatal("must be deterministic")
	}
}
