package utils

import (
	"os"
	"sync"
	"encoding/json"
	"github.com/Olling/Enrolld/config"
)


var (
	SyncOutputMutex		sync.Mutex
	SyncGetInventoryMutex	sync.Mutex
	Overwrites		map[string]Overwrite
)


type KeyValue struct {
	Key			string
	Value			string
}


type Overwrite struct {
	ServerIDRegexp		string
	InventoriesRegexp	string
	PropertiesRegexp	KeyValue
	Inventories		[]string
	Properties		map[string]string
}


func (server Server) Exist() bool {
	_, existsErr := os.Stat(config.Configuration.FileBackendDirectory + "/" + server.ServerID)

	if os.IsNotExist(existsErr) {
		return false
	} else {
		return true
	}
}


type Server struct {
	ServerID	string
	IP		string
	LastSeen	string
	NewServer	bool `json:"NewServer,omitempty"`
	Groups		[]string
	Properties	map[string]string
}

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
