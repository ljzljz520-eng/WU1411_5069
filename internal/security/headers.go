package security

import (
	"net/http"
	"strings"
)

type HeaderPolicy struct {
	StrictTransport    bool
	FrameDeny          bool
	ContentTypeNoSniff bool
	ReferrerPolicy     string
}

func DefaultHeaderPolicy() HeaderPolicy {
	return HeaderPolicy{StrictTransport: true, FrameDeny: true, ContentTypeNoSniff: true, ReferrerPolicy: "no-referrer"}
}

func ApplyHeaders(header http.Header, policy HeaderPolicy) {
	if header == nil {
		return
	}
	if policy.StrictTransport {
		header.Set("Strict-Transport-Security", "max-age=31536000")
	}
	if policy.FrameDeny {
		header.Set("X-Frame-Options", "DENY")
	}
	if policy.ContentTypeNoSniff {
		header.Set("X-Content-Type-Options", "nosniff")
	}
	if strings.TrimSpace(policy.ReferrerPolicy) != "" {
		header.Set("Referrer-Policy", policy.ReferrerPolicy)
	}
}

func ValidRequestID(value string) bool {
	if len(value) < 1 || len(value) > 80 {
		return false
	}
	for _, char := range value {
		if !(char == '-' || char == '_' || char == '.' || char >= '0' && char <= '9' || char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z') {
			return false
		}
	}
	return true
}

func CanonicalRequestID(value string) string {
	value = strings.TrimSpace(value)
	if !ValidRequestID(value) {
		return ""
	}
	return value
}
