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

//type AnsibleAddons struct {
//	Addons			[]AnsibleAddon
//}

type Overwrite struct {
	ID			string
	ServerIDRegexp		string
	InventoriesRegexp	string
	PropertiesRegexp	KeyValue
	Inventories		[]string
	Properties		map[string]string
}


func (server ServerInfo) Exist() bool {
	_, existsErr := os.Stat(config.Configuration.FileBackendDirectory + "/" + server.ServerID)

	if os.IsNotExist(existsErr) {
		return false
	} else {
		return true
	}
}


//#type Inventory struct {
//#
//#	Hello struct {
//#		Hosts []string `json:"hosts"`
//#	} `json:"Hello"`
//#	World struct {
//#		Hosts []string `json:"hosts"`
//#	} `json:"World"`
//#	Meta struct {
//#		Hostvars struct {
//#			Localhost struct {
//#				Domain      string `json:"domain"`
//#				Environment string `json:"environment"`
//#			} `json:"localhost"`
//#		} `json:"hostvars"`
//#	} `json:"_meta"`
//#}



//type Inventory struct {
//	Groups		map[string]Group
//	Meta		MetaData `json:"_meta"`
//}

type ServerInfo struct {
	ServerID          string
	IP                string
	LastSeen          string
	NewServer         bool `json:"NewServer,omitempty"`
	Inventories       []string
	Properties map[string]string
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
