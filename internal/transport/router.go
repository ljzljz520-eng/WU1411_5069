package transport

import (
	"net/http"
	"strings"

	"example.com/temporary-share-gateway/internal/gateway"
	"example.com/temporary-share-gateway/internal/metrics"
	"example.com/temporary-share-gateway/internal/security"
)

type Router struct {
	mux     *http.ServeMux
	metrics *metrics.Counter
}

func NewRouter(counters *metrics.Counter) *Router {
	if counters == nil {
		counters = metrics.New()
	}
	return &Router{mux: http.NewServeMux(), metrics: counters}
}

func (r *Router) MountShare(handler *gateway.Handler) {
	if r == nil || r.mux == nil || handler == nil {
		return
	}
	r.mux.Handle("/share/", r.withSecurity(handler))
}

func (r *Router) MountAdmin(handler http.Handler) {
	if r == nil || r.mux == nil || handler == nil {
		return
	}
	r.mux.Handle("/admin/", r.withSecurity(handler))
}

func (r *Router) MountHealth(ready func() error) {
	if r == nil || r.mux == nil {
		return
	}
	r.mux.Handle("/healthz", r.withSecurity(HealthHandler(ready)))
}

func (r *Router) MountMetrics() {
	if r == nil || r.mux == nil {
		return
	}
	r.mux.Handle("/metrics", MetricsHandler(r.metrics.Prometheus))
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if r == nil || r.mux == nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	r.mux.ServeHTTP(writer, request)
}

func (r *Router) withSecurity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		security.ApplyHeaders(writer.Header(), security.DefaultHeaderPolicy())
		if strings.Contains(request.URL.Path, "\\") {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
