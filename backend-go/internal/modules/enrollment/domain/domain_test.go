package domain

import "testing"

func TestOrderPublicStatus(t *testing.T) {
	cases := map[string]string{
		"paid":      "completed",
		"failed":    "refunded",
		"cancelled": "refunded",
		"pending":   "pending",
		"unknown":   "pending",
	}
	for in, want := range cases {
		if got := (Order{Status: in}).PublicStatus(); got != want {
			t.Errorf("PublicStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestThumbnailColorDeterministic(t *testing.T) {
	if ThumbnailColor(7) != ThumbnailColor(7) {
		t.Fatal("must be deterministic")
	}
}
