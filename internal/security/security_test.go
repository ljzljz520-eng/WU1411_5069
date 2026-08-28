package security

import "testing"

func TestTokenRedactionAndHashing(t *testing.T) {
	if HashToken("secret") == HashToken("other") {
		t.Fatal("hash collision")
	}
	if RedactToken("abcdef") != "ab...ef" {
		t.Fatal("redaction mismatch")
	}
	if SafeDetail("line\nvalue", 20) != "line value" {
		t.Fatal("detail was not normalized")
	}
}

func TestSensitiveHeaderClassification(t *testing.T) {
	if !IsSensitiveHeader("Authorization") || IsSensitiveHeader("Accept") {
		t.Fatal("header classification mismatch")
	}
}
