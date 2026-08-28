package transport

import (
	"net/http"
)

func HealthHandler(ready func() error) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if ready != nil {
			if err := ready(); err != nil {
				writer.WriteHeader(http.StatusServiceUnavailable)
				_, _ = writer.Write([]byte("not ready"))
				return
			}
		}
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("ok"))
	})
}

func MetricsHandler(render func() string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; version=0.0.4")
		if render != nil {
			_, _ = writer.Write([]byte(render()))
		}
	})
}
