package update

import "testing"

func TestNormalizeSemverString(t *testing.T) {
	got, err := NormalizeSemverString("v0.0.10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "0.0.10" {
		t.Fatalf("got %q", got)
	}
}

func TestCompareSemver(t *testing.T) {
	a, _ := ParseSemver("0.0.2")
	b, _ := ParseSemver("0.0.10")
	if CompareSemver(a, b) >= 0 {
		t.Fatalf("expected 0.0.2 < 0.0.10")
	}
}
