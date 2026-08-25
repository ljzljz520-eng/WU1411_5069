package model

const (
	StatusActive  = "active"
	StatusRevoked = "revoked"
	StatusExpired = "expired"
	OutcomeAllow  = "allow"
	OutcomeDeny   = "deny"
)

func ValidStatus(value string) bool {
	switch value {
	case StatusActive, StatusRevoked, StatusExpired:
		return true
	default:
		return false
	}
}

func ValidOutcome(value string) bool {
	return value == OutcomeAllow || value == OutcomeDeny
}
