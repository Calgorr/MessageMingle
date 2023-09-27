package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	MethodCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_server_method_count",
			Help: "Count of successful/failed RPC calls by method",
		},
		[]string{"method", "status", "podID"},
	)

	MethodDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "grpc_server_method_duration",
		Help: "Latency of RPC calls by method",
	},
		[]string{"method"},
	)

	ActiveSubscribers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "grpc_server_active_subscribers",
			Help: "Total active subscriptions",
		},
	)
)

func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
