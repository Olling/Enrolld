package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metric definintions
var (
	ServersAdded = promauto.NewCounter(prometheus.CounterOpts{
		Subsystem: "enrolld",
		Name:      "servers_added_total",
		Help:      "The total number of added servers",
	})

	ServersUpdated = promauto.NewCounter(prometheus.CounterOpts{
		Subsystem: "enrolld",
		Name:      "servers_updated_total",
		Help:      "The total number of updated servers",
	})

	ServersDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Subsystem: "enrolld",
		Name:      "servers_deleted_total",
		Help:      "The total number of deleted servers",
	})

	ServersInInventory prometheus.GaugeFunc
	DataUsage prometheus.GaugeFunc
	JobQueueCount prometheus.GaugeFunc
	WorkingQueueCount prometheus.GaugeFunc
)
