package gateway

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"example.com/temporary-share-gateway/internal/metrics"
	"example.com/temporary-share-gateway/internal/model"
	"example.com/temporary-share-gateway/internal/security"
	"example.com/temporary-share-gateway/internal/share"
)

type ResourceProvider interface {
	Read(resourceID string) (string, bool)
}

type StaticResources map[string]string

func (r StaticResources) Exists(resourceID string) bool {
	_, ok := r[resourceID]
	return ok
}

func (r StaticResources) Read(resourceID string) (string, bool) {
	value, ok := r[resourceID]
	return value, ok
}

type Handler struct {
	service   *share.ShareService
	auditor   *security.Auditor
	resources ResourceProvider
	metrics   *metrics.Counter
	requireID bool
	limiter   *share.Limiter
}

func NewHandler(service *share.ShareService, auditor *security.Auditor, resources ResourceProvider, counters *metrics.Counter, requireID bool) *Handler {
	if resources == nil {
		resources = StaticResources{}
	}
	if counters == nil {
		counters = metrics.New()
	}
	limiter := share.NewLimiter(service, 100, time.Minute)
	return &Handler{service: service, auditor: auditor, resources: resources, metrics: counters, requireID: requireID, limiter: limiter}
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	parsed, err := ParseRequest(request)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err.Error(), "")
		h.metrics.Inc("requests_invalid")
		return
	}
	session := security.NewSession(parsed.RequestID, parsed.Token, request)
	if !session.Valid(false) && parsed.Token != "" {
		writeError(writer, http.StatusBadRequest, "invalid_session", "request session is invalid", parsed.RequestID)
		return
	}
	request = request.WithContext(security.WithSession(request.Context(), session))
	if h.requireID && parsed.RequestID == "" {
		writeError(writer, http.StatusBadRequest, "request_id_required", "request id is required", "")
		h.metrics.Inc("requests_invalid")
		return
	}
	if h.limiter != nil && parsed.Token != "" {
		allowed, window, rateErr := h.limiter.Allow(request.Context(), parsed.Token)
		if rateErr != nil {
			writeError(writer, http.StatusInternalServerError, "rate_limit_failed", "request could not be completed", parsed.RequestID)
			return
		}
		if !allowed {
			writer.Header().Set("Retry-After", formatInt(int(time.Minute.Seconds())))
			writeError(writer, http.StatusTooManyRequests, "rate_limited", "request rate exceeded", parsed.RequestID)
			return
		}
		h.metrics.Add("rate_remaining", int64(h.limiter.Remaining(window)))
	}
	ctx := request.Context()
	record, err := h.service.Authorize(ctx, parsed.Token, parsed.RequestID)
	if err != nil {
		h.metrics.Inc("requests_denied")
		h.audit(ctx, model.AccessEvent{Token: parsed.Token, ResourceID: parsed.Resource, Outcome: model.OutcomeDeny, Reason: err.Error(), At: record.LastUsedAt, RequestID: parsed.RequestID}, parsed.Token)
		var denial share.Denial
		if errors.As(err, &denial) {
			writeError(writer, statusCode(denial.Status), denialCode(denial), denial.Error(), parsed.RequestID)
			return
		}
		writeError(writer, http.StatusInternalServerError, "internal_error", "request could not be completed", parsed.RequestID)
		return
	}
	body, ok := h.resources.Read(parsed.Resource)
	if !ok {
		h.metrics.Inc("resources_missing")
		h.audit(ctx, model.AccessEvent{Token: parsed.Token, ResourceID: parsed.Resource, Outcome: model.OutcomeDeny, Reason: share.ErrResourceGone.Error(), RequestID: parsed.RequestID}, parsed.Token)
		writeError(writer, http.StatusNotFound, "resource_missing", share.ErrResourceGone.Error(), parsed.RequestID)
		return
	}
	h.metrics.Inc("requests_allowed")
	h.audit(ctx, model.AccessEvent{Token: parsed.Token, ResourceID: record.ResourceID, Outcome: model.OutcomeAllow, Reason: "share consumed", RequestID: parsed.RequestID}, parsed.Token)
	if request.Method == http.MethodHead {
		writer.WriteHeader(http.StatusOK)
		return
	}
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.Header().Set("X-Share-Remaining", strings.TrimSpace(strings.Join([]string{formatInt(record.Remaining)}, "")))
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(body))
}

func (h *Handler) audit(ctx context.Context, event model.AccessEvent, token string) {
	if h.auditor == nil {
		return
	}
	if event.At.IsZero() {
		event.Reason = security.SafeDetail(event.Reason, 160)
	}
	_ = h.auditor.Record(ctx, event, token)
}

func denialCode(denial share.Denial) string {
	switch denial.Cause {
	case share.ErrTokenMissing:
		return "token_missing"
	case share.ErrTokenInvalid:
		return "token_invalid"
	case share.ErrExpired:
		return "token_expired"
	case share.ErrExhausted:
		return "token_exhausted"
	case share.ErrRevoked:
		return "token_revoked"
	default:
		return "request_denied"
	}
}

func formatInt(value int) string {
	if value == 0 {
		return "0"
	}
	if value < 0 {
		return "-1"
	}
	result := make([]byte, 0, 12)
	for value > 0 {
		result = append([]byte{byte('0' + value%10)}, result...)
		value /= 10
	}
	return string(result)
}
