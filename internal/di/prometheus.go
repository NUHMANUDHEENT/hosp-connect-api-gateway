package di

import "github.com/prometheus/client_golang/prometheus"

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of requests processed",
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(requestCount)
}
