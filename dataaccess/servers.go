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
	slog.PrintDebug("Getting server count")

	servers, err := GetServers()
	if err != nil {
		return 0
	}

	slog.PrintTrace("Number of servers:", len(servers))

	return float64(len(servers))
}

func GetServers() ([]objects.Server, error) {
	slog.PrintDebug("Get Servers")
	slog.PrintTrace("Get servers with the following Overwrites:", Overwrites)

	switch Backend {
		case "file":
			return fileio.GetServers(Overwrites)
	}

	return nil, errors.New("Selected backend is unknown")
}

func GetServer(serverID string) (server objects.Server, err error) {
	slog.PrintDebug("Get Server")
	slog.PrintTrace("Get server with the following ServerID and Overwrites:", serverID, Overwrites)

	switch Backend {
		case "file":
			return fileio.GetServer(serverID, Overwrites)
	}

	return server, errors.New("Selected backend is unknown")
}

func ServerExist(server objects.Server) bool {
	slog.PrintDebug("Checking if server exists")
	slog.PrintTrace("Checking if the following server exists:", server)

	switch Backend {
		case "file":
			return fileio.ServerExist(server.ServerID)
	}

	return false
}

func RemoveServer(serverID string) error {
	slog.PrintDebug("Removing server")
	slog.PrintTrace("Removing server with the following server ID:", serverID)

	switch Backend {
		case "file":
			err := fileio.RemoveServer(serverID)
			if err == nil {
				metrics.ServersDeleted.Inc()
			} else {
				slog.PrintTrace("Removing server gave the following error:", serverID, err)
			}
			return err
	}
	return errors.New("Selected backend is unknown")
}

func GetFilteredServersList(groups []string, properties map[string]string) ([]objects.Server, error) {
	slog.PrintDebug("Get filtered server list")
	slog.PrintTrace("Creating filtered server list form the following information:", groups, properties)

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
	slog.PrintDebug("Updating server")
	slog.PrintTrace("Updating server with the following information:", isNewServer, server)

	switch Backend {
		case "file":
			return fileio.UpdateServer(server, isNewServer)
	}

	return errors.New("Selected backend is unknown")
}
