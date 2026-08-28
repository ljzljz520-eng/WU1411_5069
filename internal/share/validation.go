package share

import (
	"net/http"
	"strings"
)

const TokenHeader = "X-Share-Token"

func TokenFromHeader(header http.Header) string {
	if header == nil {
		return ""
	}
	return strings.TrimSpace(header.Get(TokenHeader))
}

func RequestIDFromHeader(header http.Header) string {
	if header == nil {
		return ""
	}
	return strings.TrimSpace(header.Get("X-Request-ID"))
}

func NormalizeResourceID(resourceID string) string {
	return strings.Trim(strings.TrimSpace(resourceID), "/")
}

func IsSafeResourceID(resourceID string) bool {
	if resourceID == "" || strings.Contains(resourceID, "..") {
		return false
	}
	for _, char := range resourceID {
		if char == '\\' || char == '\x00' {
			return false
		}
	}
	return true
}
