package output

import (
	"time"
	"errors"
	"strings"
	"io/ioutil"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/fileio"
	"github.com/Olling/Enrolld/config"
)


func CategorizeGroups(servers []utils.Server) ([]string, map[string][]utils.Server) {
	keys := make([]string, 0)
	results := make(map[string][]utils.Server)

	for _,server := range servers {
		for _, group := range server.Groups {
			if results[group] != nil {
				results[group] = append(results[group], server)
			} else {
				keys = append(keys, group)
				results[group] = []utils.Server{server}
			}
		}
	}

	return keys, results
}


func GetInventoryInJSON(servers []utils.Server) (json string, err error) {
	type Group struct {
		Hosts		[]string `json:"hosts"`
	}

	type Meta struct {
		Hostvars	map[string]map[string]string `json:"hostvars"`
	}

	var inventory = make(map[string]interface{})
	hostvars := make(map[string]map[string]string)
	inventory["_meta"] = Meta{Hostvars: hostvars}

	for _, server := range servers {
		for _,serverInventory := range server.Groups {
			if _,ok := inventory[serverInventory]; ok {
				group := inventory[serverInventory].(Group)
				group.Hosts = append(group.Hosts, server.ServerID)
				inventory[serverInventory] = group

				continue
			}

			inventory[serverInventory] = Group{Hosts: []string{server.ServerID}}
		}

		if server.Properties == nil {
			continue
		}

		meta := inventory["_meta"].(Meta)
		meta.Hostvars[server.ServerID] = server.Properties
		inventory["_meta"] = meta

		continue
	}

	json,err = utils.StructToJson(inventory)

	return json, err
}


func GetServer(serverID string) (server utils.Server, err error) {
	err = fileio.LoadFromFile(&server, config.Configuration.FileBackendDirectory + "/" + serverID)

	if err != nil {
		return server, err
	}

	fileio.AddOverwrites(&server)

	layout := "2006-01-02 15:04:05.999999999 -0700 MST"

	if strings.Contains(server.LastSeen, "m=") {
		server.LastSeen = strings.Split(server.LastSeen, " m=")[0]
	}

	date, err := time.Parse(layout, server.LastSeen)

	if err == nil {
		return server, err
	}

	date = date.Add(time.Minute * time.Duration(config.Configuration.MaxAgeInMinutes))

	if date.After(time.Now()) {
		return server,nil
	}

	return server, errors.New("Server was beyond max age")
}


func GetServerCount() float64 {
	filelist, filelisterr := ioutil.ReadDir(config.Configuration.FileBackendDirectory)
	if filelisterr != nil {
		return 0
	}

	return float64(len(filelist))
}


func GetServers() ([]utils.Server, error) {
	var inventory []utils.Server

	filelist, filelisterr := ioutil.ReadDir(config.Configuration.FileBackendDirectory)
	if filelisterr != nil {
		slog.PrintError("Failed to get inventory:", filelisterr)
		return nil, filelisterr
	}

	utils.SyncGetInventoryMutex.Lock()
	defer utils.SyncGetInventoryMutex.Unlock()

	for _, child := range filelist {
		if child.IsDir() == false {
			server, err := GetServer(child.Name())

			if err != nil && err.Error() != "Server was beyond max age" {
				slog.PrintError("Error while reading file", config.Configuration.FileBackendDirectory + "/" + child.Name(), "Reason:", err)
				continue
			}

			inventory = append(inventory, server)
		}
	}
	return inventory, nil
}

func GetFilteredServersList(groups []string, properties map[string]string) ([]utils.Server, error) {
	servers, err := GetServers()
	var filteredServers []utils.Server

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
