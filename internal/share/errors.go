package share

import "errors"

var (
	ErrTokenMissing = errors.New("share token is missing")
	ErrTokenInvalid = errors.New("share token is invalid")
	ErrExpired      = errors.New("share token is expired")
	ErrExhausted    = errors.New("share token has no remaining uses")
	ErrRevoked      = errors.New("share token is revoked")
	ErrResourceGone = errors.New("shared resource is unavailable")
)

type Denial struct {
	Cause  error
	Status int
}

func (d Denial) Error() string { return d.Cause.Error() }

func (d Denial) Unwrap() error { return d.Cause }

func denial(cause error, status int) error {
	return Denial{Cause: cause, Status: status}
}
