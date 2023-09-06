package trace

import (
	"fmt"
	"log"
	"net/http"

	prm "therealbroker/internal/prometheus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func PrometheusServerStart() {
	prometheus.MustRegister(prm.MethodDuration)
	prometheus.MustRegister(prm.MethodCount)
	prometheus.MustRegister(prm.ActiveSubscribers)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatalf("Prometheus server failed: %v", http.ListenAndServe(fmt.Sprintf(":%d", 9091), nil))
}
