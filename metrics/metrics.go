package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metric definintions
var (
	ServerAddedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Subsystem: "enrolld",
		Name:      "server_added_total",
		Help:      "The total number of added servers",
	})

	ServerUpdatedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Subsystem: "enrolld",
		Name:      "server_updated_total",
		Help:      "The total number of updated servers",
	})
)
