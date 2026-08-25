package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/temporary-share-gateway/internal/metrics"
)

func TestHealthAndMetricsHandlers(t *testing.T) {
	health := httptest.NewRecorder()
	HealthHandler(func() error { return nil }).ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK || health.Body.String() != "ok" {
		t.Fatalf("health = %d %q", health.Code, health.Body.String())
	}
	counter := metrics.New()
	counter.Inc("requests_total")
	metricsResponse := httptest.NewRecorder()
	MetricsHandler(counter.Prometheus).ServeHTTP(metricsResponse, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if metricsResponse.Code != http.StatusOK || metricsResponse.Body.Len() == 0 {
		t.Fatal("metrics response is empty")
	}
}

func TestMiddlewareRejectsDoubleSlash(t *testing.T) {
	handler := WithDefaults(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusNoContent) }), metrics.New())
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/share//asset", nil))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}
