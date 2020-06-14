package utils

import (
	"os"
	"fmt"
	"net"
	"errors"
	"regexp"
	"os/exec"
	"strings"
	"encoding/json"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enrolld/dataaccess/config"
)

//var (
//	SyncOutputMutex		sync.Mutex
//	SyncGetInventoryMutex	sync.Mutex
//	SyncActiveMutex		sync.Mutex
//	Overwrites		map[string]Overwrite
//	Scripts			map[string]Script
//	ActiveServers		map[string]time.Time
//)
//
//type KeyValue struct {
//	Key			string
//	Value			string
//}
//
//type Overwrite struct {
//	ServerIDRegexp		string
//	GroupRegexp		string
//	PropertiesRegexp	KeyValue
//	Groups			[]string
//	Properties		map[string]string
//}
//
//type Script struct {
//	Description	string
//	Path		string
//	Timeout		int
//}
//
//type Server struct {
//	ServerID	string
//	IP		string
//	LastSeen	string
//	NewServer	bool `json:"NewServer,omitempty"`
//	Groups		[]string
//	Properties	map[string]string
//}

func StructToJson(s interface{}) (string, error) {
	bytes, marshalErr := json.MarshalIndent(s, "", "\t")
	return string(bytes), marshalErr
}

func StructFromJson(input []byte, output interface{}) error {
	return json.Unmarshal(input, &output)
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

func Notification(subject string, message string, server objects.Server) {
	binary, err := exec.LookPath(config.Configuration.NotificationScriptPath)
	if err != nil {
		slog.PrintError("Could not find the notification script in the given path", config.Configuration.NotificationScriptPath, err)
	}
	cmd := exec.Command(binary)

	env := os.Environ()
	env = append(env, fmt.Sprintf("SUBJECT=%s", subject))
	env = append(env, fmt.Sprintf("MESSAGE=%s", message))

	env = append(env, fmt.Sprintf("SERVER_ID=%s", server.ServerID))
	env = append(env, fmt.Sprintf("SERVER_IP=%s", server.IP))
	env = append(env, fmt.Sprintf("SERVER_PROPERTIES=%s", server.Properties))
	env = append(env, fmt.Sprintf("SERVER_INVENTORIES=%s", server.Groups))
	env = append(env, fmt.Sprintf("SERVER_LASTSEEN=%s", server.LastSeen))

	cmd.Env = env

	startErr := cmd.Start()
	if startErr != nil {
		slog.PrintError("Could not send notification", startErr)
	}

	cmd.Wait()
}

func ValidInput(input string) bool {
	matched,_ := regexp.MatchString(input, "^[a-zA-Z0-9_-.]*$")
	return !matched
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

	json,err = StructToJson(inventory)

	return json, err
}

func VerifyFQDN(serverid string, requestIP string) (string, error) {
	var fqdn string

	if serverid == "" {
		addresses, err := net.LookupAddr(requestIP)
		if err == nil && len(addresses) >= 1 {
			addr := addresses[0]
			if addr != "" {
				addr = strings.TrimSuffix(addr, ".")
				slog.PrintInfo("FQDN was empty (" + requestIP + ") but the IP had the following name: \"" + addr + "\"")
				serverid = addr
			}
		} else {
			return "", errors.New("FQDN is empty")
		}
	}

	if m, _ := regexp.MatchString("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$", serverid); !m {
		slog.PrintError("Received FQDN with illegal characters: \"", fqdn, "\" (", requestIP, ")")
		return "", errors.New("Received FQDN with illegal characters: \"" + fqdn + "\" (" + requestIP + ")")
	}

	if len(strings.Split(serverid, ".")) < 3 {
		addresses, err := net.LookupAddr(requestIP)
		if err == nil && len(addresses) >= 1 {
			addr := addresses[0]
			if addr != "" {
				addr = strings.TrimSuffix(addr, ".")
				slog.PrintDebug("Server \"" + serverid + "\"'s domain looks wrong - Replacing it with \"" + addr + "\"")
				serverid = addr
			}
		} else {
			slog.PrintError("Server \"" + serverid + "\"'s domain looks wrong, but no suitable name was found to replace it")
		}
	}
	return serverid, nil
}

