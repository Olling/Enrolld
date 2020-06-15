package dataaccess

import (
	"errors"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/metrics"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enrolld/dataaccess/fileio"
)

func GetServerCount() float64 {
	servers, err := GetServers()
	if err != nil {
		return 0
	}

	return float64(len(servers))
}

func GetServers() ([]objects.Server, error) {
	switch Backend {
		case "file":
			return fileio.GetServers(Overwrites)
	}

	return nil, errors.New("Selected backend is unknown")
}

func GetServer(serverID string) (server objects.Server, err error) {
	switch Backend {
		case "file":
			return fileio.GetServer(serverID, Overwrites)
	}

	return server, errors.New("Selected backend is unknown")
}

func ServerExist(server objects.Server) bool {
	switch Backend {
		case "file":
			return fileio.ServerExist(server.ServerID)
	}

	return false
}

func RemoveServer(serverID string) error {
	switch Backend {
		case "file":
			err := fileio.RemoveServer(serverID)
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

func UpdateServer(server objects.Server, isNewServer bool) error {
	switch Backend {
		case "file":
			slog.PrintDebug("Updating server on disk")
			return fileio.UpdateServer(server, isNewServer)
		default:
			return errors.New("Selected backend is unknown")
	}
}
