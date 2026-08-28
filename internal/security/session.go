package security

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type Session struct {
	RequestID string
	TokenHash string
	RemoteIP  string
	mu        sync.RWMutex
	values    map[string]string
}

func NewSession(requestID, token string, request *http.Request) *Session {
	remote := ""
	if request != nil {
		remote = request.RemoteAddr
	}
	return &Session{RequestID: CanonicalRequestID(requestID), TokenHash: HashToken(token), RemoteIP: remote, values: make(map[string]string)}
}

func (s *Session) Valid(requireID bool) bool {
	if s == nil {
		return false
	}
	if requireID && s.RequestID == "" {
		return false
	}
	return s.TokenHash != HashToken("")
}

func (s *Session) Set(key, value string) error {
	if s == nil || strings.TrimSpace(key) == "" {
		return fmt.Errorf("session key is required")
	}
	s.mu.Lock()
	s.values[key] = SafeDetail(value, 200)
	s.mu.Unlock()
	return nil
}

func (s *Session) Get(key string) string {
	if s == nil {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.values[key]
}

func (s *Session) Fields() map[string]string {
	if s == nil {
		return map[string]string{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	fields := map[string]string{"request_id": s.RequestID, "token_hash": s.TokenHash, "remote_ip": s.RemoteIP}
	for key, value := range s.values {
		fields[key] = value
	}
	return fields
}

type contextKey string

const sessionKey contextKey = "share-session"

func WithSession(ctx context.Context, session *Session) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, sessionKey, session)
}

func SessionFromContext(ctx context.Context) *Session {
	if ctx == nil {
		return nil
	}
	session, _ := ctx.Value(sessionKey).(*Session)
	return session
}
