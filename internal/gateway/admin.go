package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example.com/temporary-share-gateway/internal/model"
	"example.com/temporary-share-gateway/internal/persist"
	"example.com/temporary-share-gateway/internal/share"
)

type AdminHandler struct {
	registry *share.Registry
	store    *persist.Store
	now      func() time.Time
}

func NewAdminHandler(registry *share.Registry, store *persist.Store, now func() time.Time) *AdminHandler {
	if now == nil {
		now = func() time.Time { return time.Unix(0, 0).UTC() }
	}
	return &AdminHandler{registry: registry, store: store, now: now}
}

func (h *AdminHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if h == nil || h.registry == nil {
		writeError(writer, http.StatusServiceUnavailable, "admin_unavailable", "admin service is unavailable", "")
		return
	}
	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/admin/shares":
		h.list(writer, request)
	case request.Method == http.MethodPost && strings.HasSuffix(request.URL.Path, "/revoke"):
		h.revoke(writer, request)
	case request.Method == http.MethodPost && strings.HasSuffix(request.URL.Path, "/rotate"):
		h.rotate(writer, request)
	case request.Method == http.MethodGet && request.URL.Path == "/admin/export":
		h.export(writer)
	default:
		writeError(writer, http.StatusNotFound, "admin_route_missing", "admin route is unavailable", "")
	}
}

func (h *AdminHandler) list(writer http.ResponseWriter, request *http.Request) {
	summaries, err := h.registry.List(request.Context())
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "admin_list_failed", err.Error(), "")
		return
	}
	writeJSON(writer, http.StatusOK, Response{Code: "share_list", ResourceID: strconv.Itoa(len(summaries))})
	_ = json.NewEncoder(writer).Encode(summaries)
}

func (h *AdminHandler) revoke(writer http.ResponseWriter, request *http.Request) {
	token := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/admin/shares/"), "/revoke")
	if token == "" {
		writeError(writer, http.StatusBadRequest, "token_missing", "token is required", "")
		return
	}
	if err := h.registry.Revoke(request.Context(), token); err != nil {
		writeError(writer, http.StatusNotFound, "revoke_failed", err.Error(), "")
		return
	}
	writeJSON(writer, http.StatusOK, Response{Code: "revoked", RequestID: token})
}

func (h *AdminHandler) rotate(writer http.ResponseWriter, request *http.Request) {
	token := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/admin/shares/"), "/rotate")
	grant := model.TokenGrant{}
	if err := DecodeJSONBody(request, &grant, 1<<16); err != nil {
		writeError(writer, http.StatusBadRequest, "rotate_payload_invalid", err.Error(), "")
		return
	}
	if grant.ExpiresAt.IsZero() {
		grant.ExpiresAt = h.now().Add(time.Hour)
	}
	record, err := h.registry.Rotate(request.Context(), token, grant.ExpiresAt, grant.Uses)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "rotate_failed", err.Error(), "")
		return
	}
	writeJSON(writer, http.StatusOK, Response{Code: "rotated", ResourceID: record.ResourceID, Remaining: record.Remaining})
}

func (h *AdminHandler) export(writer http.ResponseWriter) {
	if h.store == nil {
		writeError(writer, http.StatusServiceUnavailable, "store_unavailable", "store is unavailable", "")
		return
	}
	data, err := h.store.SnapshotJSON(h.now())
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "export_failed", fmt.Sprintf("export failed: %v", err), "")
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(data)
}

func (h *AdminHandler) Context() context.Context { return context.Background() }
