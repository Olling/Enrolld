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

func CategorizeInventories(inventories []utils.ServerInfo) ([]string, map[string][]utils.ServerInfo) {
	keys := make([]string, 0)
	results := make(map[string][]utils.ServerInfo)

	for _, inventory := range inventories {
		for _, foundInventoryName := range inventory.Inventories {
			if results[foundInventoryName] != nil {
				results[foundInventoryName] = append(results[foundInventoryName], inventory)
			} else {
				keys = append(keys, foundInventoryName)
				results[foundInventoryName] = []utils.ServerInfo{inventory}
			}
		}
	}

	return keys, results
}


func GetInventoryInJSON(servers []utils.ServerInfo) (json string, err error) {
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
		for _,serverInventory := range server.Inventories {
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


func GetServer(serverID string) (server utils.ServerInfo, err error) {
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


func GetInventoryCount() float64 {
	filelist, filelisterr := ioutil.ReadDir(config.Configuration.FileBackendDirectory)
	if filelisterr != nil {
		return 0
	}

	return float64(len(filelist))
}


func GetInventory() ([]utils.ServerInfo, error) {
	var inventory []utils.ServerInfo

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

func GetFilteredInventory(inventories []string, properties map[string]string) ([]utils.ServerInfo, error) {
	inventory, err := GetInventory()
	var filteredInventory []utils.ServerInfo

	if err != nil {
		return inventory, err
	}

	for _, server := range inventory {
		if len(inventories) != 0 {
			for _,inventory := range inventories {
				if !utils.StringExistsInArray(server.Inventories, inventory) {
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

		filteredInventory = append(filteredInventory, server)
	}

	return filteredInventory, nil
}


