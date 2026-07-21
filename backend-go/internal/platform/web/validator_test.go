package web

import "testing"

func TestValidatorCollectsFirstErrorPerKey(t *testing.T) {
	v := NewValidator()
	v.Check(false, "email", "must be provided")
	v.Check(false, "email", "second message ignored")
	if v.Valid() {
		t.Fatal("expected invalid")
	}
	if got := v.Errors["email"]; got != "must be provided" {
		t.Fatalf("got %q, want first message", got)
	}
}

func TestEmailRX(t *testing.T) {
	// The pattern mirrors the original backend's EMAIL_RX, which permits
	// single-label domains (e.g. "a@b").
	valid := []string{"a@b.co", "user.name+tag@example.com", "a@b"}
	invalid := []string{"", "no-at", "a@", "@b.co"}
	for _, e := range valid {
		if !Matches(e, EmailRX) {
			t.Errorf("%q should be valid", e)
		}
	}
	for _, e := range invalid {
		if Matches(e, EmailRX) {
			t.Errorf("%q should be invalid", e)
		}
	}
}

func TestPermitted(t *testing.T) {
	if !Permitted("uz", "uz", "ru", "en") {
		t.Error("uz should be permitted")
	}
	if Permitted("fr", "uz", "ru", "en") {
		t.Error("fr should not be permitted")
	}
}

func TestParseIDList(t *testing.T) {
	got := ParseIDList("1, 2 ,x,3,-4,0")
	want := []int64{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
