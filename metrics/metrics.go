package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metric definintions
var (
	ServerAddedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "enrolld_server_added_total",
		Help: "The total number of added servers",
	})
)

var (
	ServerUpdatedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "enrolld_server_updated_total",
		Help: "The total number of updated servers",
	})
)

// Init initializes and registers all prometheus metrics to be exposed
func Init() {
	prometheus.MustRegister(ServerAddedCounter)
	prometheus.MustRegister(ServerUpdatedCounter)
}
