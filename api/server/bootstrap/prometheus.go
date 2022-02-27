package bootstrap

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var internalPrometheusPort = 8000

var MethodCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "rpc_request_get_method_count",
		Help: "Count of each rpc call.",
	},
	[]string{"method", "status"},
)

var MethodDuration = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name:       "rpc_request_get_method_duration",
		Help:       "Measure the duration of each RPC request.",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	},
	[]string{"rpc_duration"},
)

var ActiveSubscribersGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "active_subscribers",
	Help: "Number of active subscribers",
})

func StartPrometheusServer() {
	if (prometheus.Register(MethodCount) != nil) || (prometheus.Register(MethodDuration) != nil) ||
		(prometheus.Register(ActiveSubscribersGauge) != nil) {
		log.Fatalf("Some error with prometheus")
		return
	}

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", internalPrometheusPort), nil))
}
