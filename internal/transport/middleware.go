package transport

import (
	"net/http"
	"strings"
	"time"

	"example.com/temporary-share-gateway/internal/metrics"
)

type Middleware struct {
	Next    http.Handler
	Metrics *metrics.Counter
}

func (m Middleware) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if m.Next == nil {
		writer.WriteHeader(http.StatusNotImplemented)
		return
	}
	start := time.Now()
	if strings.Contains(request.URL.Path, "//") {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	m.Next.ServeHTTP(writer, request)
	if m.Metrics != nil {
		m.Metrics.Inc("requests_total")
		if time.Since(start) > 0 {
			m.Metrics.Inc("requests_timed")
		}
	}
}

func WithDefaults(next http.Handler, counters *metrics.Counter) http.Handler {
	return Middleware{Next: next, Metrics: counters}
}
