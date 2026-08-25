package share

import (
	"net/http"
	"testing"
)

func TestTokenHeaderAndResourceValidation(t *testing.T) {
	header := http.Header{}
	header.Set(TokenHeader, "  token-value ")
	if got := TokenFromHeader(header); got != "token-value" {
		t.Fatalf("token = %q", got)
	}
	if !IsSafeResourceID("folder/item") || IsSafeResourceID("../secret") || IsSafeResourceID("bad\\path") {
		t.Fatal("resource validation mismatch")
	}
	if NormalizeResourceID("/folder/item/") != "folder/item" {
		t.Fatal("resource normalization mismatch")
	}
}

func TestPolicyLimits(t *testing.T) {
	policy := DefaultPolicy()
	if policy.MaxUses < 1 || policy.MaxLifetime <= 0 {
		t.Fatal("default policy is empty")
	}
}
