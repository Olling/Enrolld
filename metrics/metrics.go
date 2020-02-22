package metrics

import (
	"github.com/Olling/Enrolld/output"
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

	ServersInInventory = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Subsystem: "enrolld",
		Name:      "servers_in_inventory",
		Help:      "Total amount of servers in inventory",
	},
		func() float64 {
			inventory, err := output.GetInventory()
			if err != nil {
				return 0
			}
			return float64(len(inventory))
		})
)
