package output

import (
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/utils/objects"
)


func CategorizeGroups(servers []objects.Server) ([]string, map[string][]objects.Server) {
	keys := make([]string, 0)
	results := make(map[string][]objects.Server)

	for _,server := range servers {
		for _, group := range server.Groups {
			if results[group] != nil {
				results[group] = append(results[group], server)
			} else {
				keys = append(keys, group)
				results[group] = []objects.Server{server}
			}
		}
	}

	return keys, results
}


func GetInventoryInJSON(servers []objects.Server) (json string, err error) {
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
