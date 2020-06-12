package metrics

import (
	"os"
	"path/filepath"

	"github.com/Olling/Enrolld/output"
	"github.com/Olling/Enrolld/dataaccess/config"
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
	}, output.GetServerCount)

	DataUsage = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Subsystem: "enrolld",
		Name:      "data_usage",
		Help:      "Data usage on the disk (bytes)",
	},
		func() float64 {
			var size int64 = 0

			readSize := func(path string, file os.FileInfo, err error) error {
				// check if dir
				if !file.IsDir() {
					size += file.Size()
				}

				return nil
			}

			// recursive iterate
			filepath.Walk(config.Configuration.FileBackendDirectory, readSize)

			return float64(size)
		})
)
