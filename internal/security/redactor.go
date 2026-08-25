package security

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func HashToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func RedactToken(token string) string {
	trimmed := strings.TrimSpace(token)
	if len(trimmed) <= 4 {
		return "****"
	}
	return trimmed[:2] + "..." + trimmed[len(trimmed)-2:]
}

func SafeDetail(value string, max int) string {
	value = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return ' '
		}
		return r
	}, value)
	if max < 1 {
		return ""
	}
	if len(value) > max {
		return value[:max]
	}
	return value
}

func IsSensitiveHeader(name string) bool {
	switch strings.ToLower(name) {
	case "authorization", "x-share-token", "cookie", "set-cookie":
		return true
	default:
		return false
	}
}
