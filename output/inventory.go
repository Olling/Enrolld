package output

import (
	"os"
	"time"
	"errors"
	"strings"
	"io/ioutil"
	"encoding/json"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
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


func GetInventoryInJSON(inventories []utils.ServerInfo) (string, error) {
	inventoryjson := "{"

	keys, inventoryMap := CategorizeInventories(inventories)

	inventoryjson += "\n\t\"" + config.Configuration.DefaultInventoryName + "\"\t: {\n\t\"hosts\"\t: ["
	for _, inventory := range inventories {
		inventoryjson += "\"" + inventory.ServerID + "\", "
	}
	inventoryjson = strings.TrimSuffix(inventoryjson, ", ")
	inventoryjson += "]\n\t},"

	for _, key := range keys {
		inventoryjson += "\n\t\"" + key + "\"\t: {\n\t\"hosts\"\t: ["
		for _, inventory := range inventoryMap[key] {
			inventoryjson += "\"" + inventory.ServerID + "\", "
		}
		inventoryjson = strings.TrimSuffix(inventoryjson, ", ")
		inventoryjson += "]\n\t},"
	}

	inventoryjson += "\n\t\"_meta\" : {\n\t\t\"hostvars\" : {"

	for _, server := range inventories {
		if len(server.AnsibleProperties) != 0 {
			propertiesjsonbytes, err := json.Marshal(server.AnsibleProperties)
			if err != nil {
				slog.PrintError("Error in converting map to json", err)
			} else {
				propertiesjson := string(propertiesjsonbytes)
				propertiesjson = strings.TrimPrefix(propertiesjson, "{")
				propertiesjson = strings.TrimSuffix(propertiesjson, "}")
				inventoryjson += "\n\t\t\t\"" + server.ServerID + "\": {\n\t\t\t\t" + propertiesjson + "\n\t\t\t},"
			}
		}
	}

	inventoryjson = strings.TrimSuffix(inventoryjson, ",")
	inventoryjson += "\n\t\t}\n\t}\n}"

	return inventoryjson, nil
}


func GetServer(serverID string) (server utils.ServerInfo, err error) {
	file, err := os.Open(config.Configuration.Path + "/" + serverID)
	defer file.Close()
	if err != nil {
		return server,err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&server)
	file.Close()

	if err != nil {
		return server, err
	} else {
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
}


func GetInventoryCount() float64 {
	filelist, filelisterr := ioutil.ReadDir(config.Configuration.Path)
	if filelisterr != nil {
		return 0
	}

	return float64(len(filelist))
}


func GetInventory() ([]utils.ServerInfo, error) {
	var inventory []utils.ServerInfo

	filelist, filelisterr := ioutil.ReadDir(config.Configuration.Path)
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
				slog.PrintError("Error while reading file", config.Configuration.Path + "/" + child.Name(), "Reason:", err)
				continue
			}

			inventory = append(inventory, server)
		}
	}
	return inventory, nil
}

func GetFilteredInventory(ansibleInventories []string, ansibleProperties map[string]string) ([]utils.ServerInfo, error) {
	inventory, err := GetInventory()
	var filteredInventory []utils.ServerInfo

	if err != nil {
		return inventory, err
	}

	for _, server := range inventory {
		if len(ansibleInventories) != 0 {
			for _,ansibleInventory := range ansibleInventories {
				if !StringExistsInArray(server.Inventories, ansibleInventory) {
					continue
				}
			}

		}
		if len(ansibleProperties) != 0 {
			for key, value := range ansibleProperties {
				if !KeyValueExistsInMap(server.AnsibleProperties, key, value) {
					continue
				}
			}
		}

		filteredInventory = append(filteredInventory, server)
	}

	return filteredInventory, nil
}


func StringExistsInArray(array []string, required string) bool {
    for _, item := range array {
        if item == required {
            return true
        }
    }
    return false
}


func KeyValueExistsInMap(chart map[string]string, requiredKey string, requiredValue string) bool {
	if value, ok := chart[requiredKey]; ok {
		if requiredValue == value {
			return true
		}
	}
	return false
}

