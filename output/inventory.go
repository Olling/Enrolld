package output

import (
	"os"
	"time"
	"fmt"
	"strings"
	"io/ioutil"
	"encoding/json"
	"github.com/Olling/Enrolld/config"
	"github.com/Olling/Enrolld/utils"
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
		inventoryjson += "\"" + inventory.FQDN + "\", "
	}
	inventoryjson = strings.TrimSuffix(inventoryjson, ", ")
	inventoryjson += "]\n\t},"

	for _, key := range keys {
		inventoryjson += "\n\t\"" + key + "\"\t: {\n\t\"hosts\"\t: ["
		for _, inventory := range inventoryMap[key] {
			inventoryjson += "\"" + inventory.FQDN + "\", "
		}
		inventoryjson = strings.TrimSuffix(inventoryjson, ", ")
		inventoryjson += "]\n\t},"
	}

	inventoryjson += "\n\t\"_meta\" : {\n\t\t\"hostvars\" : {"

	for _, server := range inventories {
		if len(server.AnsibleProperties) != 0 {
			propertiesjsonbytes, err := json.Marshal(server.AnsibleProperties)
			if err != nil {
				fmt.Println("Error in converting map to json", err)
			} else {
				propertiesjson := string(propertiesjsonbytes)
				propertiesjson = strings.TrimPrefix(propertiesjson, "{")
				propertiesjson = strings.TrimSuffix(propertiesjson, "}")
				inventoryjson += "\n\t\t\t\"" + server.FQDN + "\": {\n\t\t\t\t" + propertiesjson + "\n\t\t\t},"
			}
		}
	}

	inventoryjson = strings.TrimSuffix(inventoryjson, ",")
	inventoryjson += "\n\t\t}\n\t}\n}"

	return inventoryjson, nil
}


func GetInventory() ([]utils.ServerInfo, error) {
	path :=	config.Configuration.Path
	var inventories []utils.ServerInfo

	filelist, filelisterr := ioutil.ReadDir(path)
	if filelisterr != nil {
		fmt.Println(filelisterr)
		return nil, filelisterr
	}

	utils.SyncGetInventoryMutex.Lock()
	defer utils.SyncGetInventoryMutex.Unlock()

	for _, child := range filelist {
		if child.IsDir() == false {
			file, fileerr := os.Open(path + "/" + child.Name())

			if fileerr != nil {
				fmt.Println("Error while reading file", path+"/"+child.Name(), "Reason:", fileerr)
				continue
			}

			decoder := json.NewDecoder(file)
			var inventory utils.ServerInfo
			err := decoder.Decode(&inventory)

			if err != nil {
				fmt.Println("Error while decoding file", path+"/"+child.Name(), "Reason:", err)
			} else {
				layout := "2006-01-02 15:04:05.999999999 -0700 MST"

				if strings.Contains(inventory.LastSeen, "m=") {
					inventory.LastSeen = strings.Split(inventory.LastSeen, " m=")[0]
				}

				date, parseErr := time.Parse(layout, inventory.LastSeen)

				if parseErr != nil {
					fmt.Println("Could not parse date")
					fmt.Println(parseErr)
				}

				date = date.Add(time.Minute * time.Duration(config.Configuration.MaxAgeInMinutes))

				if date.After(time.Now()) {
					inventories = append(inventories, inventory)
				}
			}
		}
	}
	return inventories, nil
}
