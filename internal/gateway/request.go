package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"example.com/temporary-share-gateway/internal/share"
)

type Request struct {
	Token     string
	RequestID string
	Resource  string
}

func ParseRequest(r *http.Request) (Request, error) {
	if r == nil {
		return Request{}, fmt.Errorf("request is nil")
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return Request{}, fmt.Errorf("method %s is not supported", r.Method)
	}
	resource := strings.TrimPrefix(r.URL.Path, "/share/")
	resource = share.NormalizeResourceID(resource)
	if !share.IsSafeResourceID(resource) {
		return Request{}, fmt.Errorf("resource path is invalid")
	}
	return Request{Token: share.TokenFromHeader(r.Header), RequestID: share.RequestIDFromHeader(r.Header), Resource: resource}, nil
}

func DecodeJSONBody(r *http.Request, target any, maxBytes int64) error {
	if r == nil || r.Body == nil {
		return fmt.Errorf("body is required")
	}
	if maxBytes < 1 {
		return fmt.Errorf("body limit is invalid")
	}
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxBytes))
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	return nil
}
