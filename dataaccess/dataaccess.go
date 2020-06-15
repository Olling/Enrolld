package dataaccess

import (
	"os"
	"fmt"
	"sync"
	"errors"
	"path/filepath"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/metrics"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enrolld/dataaccess/db"
	"github.com/Olling/Enrolld/dataaccess/config"
	"github.com/Olling/Enrolld/dataaccess/fileio"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Backend			string
	Users			map[string]objects.User
	Scripts			map[string]objects.Script
	SyncGetInventoryMutex	sync.Mutex
)

func Initialize(backend string) {
	Backend = backend

	metrics.ServersInInventory = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Subsystem: "enrolld",
		Name:      "servers_in_inventory",
		Help:      "Total amount of servers in inventory",
	}, GetServerCount)


	metrics.DataUsage = promauto.NewGaugeFunc(prometheus.GaugeOpts{
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
}

func InitializeAuthentication() {
	slog.PrintDebug("Initializing Authentication")
	err := LoadAuthentication()
	if err != nil {
		slog.PrintError("Failed to load authentication", err)
	}
}

func LoadAuthentication() error {
	switch Backend {
		case "file":
			return fileio.LoadFromFile(&Users, config.Configuration.FileBackendDirectory + "/auth.json")
		case "db":
			return db.LoadAuthentication(&Users)
	}
	return errors.New("Selected backend is unknown")
}

func RunScript(scriptPath string, server objects.Server, scriptID string, timeout int) error {
	if server.ServerID == "" {
		slog.PrintError("Failed to call", scriptID, "script - ServerID is empty!")
		return fmt.Errorf("ServerID was not given")
	}

	err := fileio.RunScript(scriptPath, server, scriptID, timeout)

	return err
}

func LoadScripts() error {
	slog.PrintDebug("Loading Scripts")
	return fileio.LoadScripts(Scripts)
}

func CheckScriptPath(path string) error {
	return fileio.CheckScriptPath(path)
}
