package metrics

import (
	"strings"
	"testing"
)

func TestCounterSnapshotAndPrometheus(t *testing.T) {
	counter := New()
	counter.Inc("requests_total")
	counter.Add("requests_total", 2)
	if counter.Value("requests_total") != 3 {
		t.Fatal("counter value mismatch")
	}
	if !strings.Contains(counter.Prometheus(), "share_gateway_requests_total 3") {
		t.Fatal("metrics output missing")
	}
}
