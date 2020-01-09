package utils

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func PrometheusMiddleware(counterName, counterDescription string) func(next http.Handler) http.HandlerFunc {

	counter := promauto.NewCounter(prometheus.CounterOpts{
		Name: counterName,
		Help: counterDescription,
	})

	return func(next http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, h *http.Request) {
			next.ServeHTTP(w, h)
			counter.Inc()
		})
	}
}
