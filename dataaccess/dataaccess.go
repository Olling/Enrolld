package dataaccess

import (
	"fmt"
	"sync"
	"time"
	"errors"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/metrics"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enrolld/dataaccess/config"
	"github.com/Olling/Enrolld/dataaccess/fileio"
)

var (
	Backend = "file"
	Users			map[string]objects.User
	Scripts			map[string]objects.Script
	Overwrites		map[string]objects.Overwrite
	ActiveServers		map[string]time.Time
	SyncGetInventoryMutex	sync.Mutex
	SyncActiveMutex		sync.Mutex
)

func GetServerCount() float64 {
	servers, err := GetServers()
	if err != nil {
		return 0
	}

	return float64(len(servers))
}

func IsServerActive(serverID string) bool {
	if ActiveServers == nil {
		return false
	}

	SyncActiveMutex.Lock()
	defer SyncActiveMutex.Unlock()

	_, exist := ActiveServers[serverID]
	return exist
}

func GetServers() ([]objects.Server, error) {
	switch Backend {
		case "file":
			return fileio.GetServersFromDisk(Overwrites)
	}

	return nil, errors.New("Selected backend is unknown")
}

func GetServer(serverID string) (server objects.Server, err error) {
	switch Backend {
		case "file":
			return fileio.GetServerFromDisk(serverID, Overwrites)
	}

	return server, errors.New("Selected backend is unknown")
}

func ServerExist(server objects.Server) bool {
	switch Backend {
		case "file":
			return fileio.ServerExistOnDisk(server.ServerID)
	}

	return false
}

func RemoveServer(serverID string) error {
	switch Backend {
		case "file":
			err := fileio.RemoveServerFromDisk(serverID)
			if err == nil {
				metrics.ServersDeleted.Inc()
			}
			return err
	}

	return errors.New("Selected backend is unknown")
}


func GetFilteredServersList(groups []string, properties map[string]string) ([]objects.Server, error) {
	servers, err := GetServers()
	var filteredServers []objects.Server

	if err != nil {
		return filteredServers, err
	}

	for _, server := range servers {
		if len(groups) != 0 {
			for _,group := range groups {
				if !utils.StringExistsInArray(server.Groups, group) {
					continue
				}
			}

		}

		if len(properties) != 0 {
			for key, value := range properties {
				if !utils.KeyValueExistsInMap(server.Properties, key, value) {
					continue
				}
			}
		}

		filteredServers = append(filteredServers, server)
	}
	return filteredServers, nil
}

func LoadAuthentication() error {
	switch Backend {
		case "file":
			slog.PrintDebug("Initializing Authentication")
			return fileio.LoadFromFile(&Users, config.Configuration.FileBackendDirectory + "/auth.json")
	}
	return errors.New("Selected backend is unknown")
}

func UpdateServer(server objects.Server, isNewServer bool) error {
	switch Backend {
		case "file":
			slog.PrintDebug("Updating server on disk")
			return fileio.UpdateServerOnDisk(server, isNewServer)
		default:
			return errors.New("Selected backend is unknown")
	}
}

func SaveOverwrites() error {
	switch Backend {
		case "file":
			slog.PrintDebug("Saving Overwrites")
			fileio.SaveOverwrites(Overwrites)
			return nil
	}
	return errors.New("Selected backend is unknown")
}

func LoadOverwrites() error {
	switch Backend {
		case "file":
			slog.PrintDebug("Loading Overwrites")
			fileio.LoadOverwrites(&Overwrites)
			return nil
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

func InitializeAuthentication() {
	slog.PrintDebug("Initializing Authentication")
	err := LoadAuthentication()
	if err != nil {
		slog.PrintError("Failed to load authentication", err)
	}
}

func MarkServerActive(server objects.Server) error {
	if ActiveServers == nil {
		ActiveServers = make(map[string]time.Time)
	}

	SyncActiveMutex.Lock()
	defer SyncActiveMutex.Unlock()

	if _, ok := ActiveServers[server.ServerID]; !ok {
		ActiveServers[server.ServerID] = time.Now()
		return nil
	}
	return errors.New("Server is already active")
}

func MarkServerInactive(server objects.Server) error {
	if ActiveServers != nil {
		SyncActiveMutex.Lock()
		defer SyncActiveMutex.Unlock()

		if _, ok := ActiveServers[server.ServerID]; ok {
			delete(ActiveServers, server.ServerID)
		}
	}

	return nil
}

func LoadScripts() error {
	switch Backend {
		case "file":
			slog.PrintDebug("Loading Scripts")
			return fileio.LoadScripts(Scripts)
	}
	return errors.New("Selected backend is unknown")
}

func CheckScriptPath(path string) error {
	return fileio.CheckScriptPath(path)
}
