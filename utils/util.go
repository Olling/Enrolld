package utils

import (
	"os"
	"sync"
	"encoding/json"
	"github.com/Olling/Enrolld/config"
)

var (
	SyncOutputMutex       sync.Mutex
	SyncGetInventoryMutex sync.Mutex
)

func (server ServerInfo) Exist() bool {
	_, existsErr := os.Stat(config.Configuration.Path + "/" + server.FQDN)

	if os.IsNotExist(existsErr) {
		return false
	} else {
		return true
	}
}

type ServerInfo struct {
	FQDN              string
	IP                string
	LastSeen          string
	NewServer         string `json:"NewServer,omitempty"`
	Inventories       []string
	AnsibleProperties map[string]string
}

func StructToJson(s interface{}) (string, error) {
	bytes, marshalErr := json.MarshalIndent(s, "", "\t")
	return string(bytes), marshalErr
}

func StructFromJson(input string, output interface{}) error {
	return json.Unmarshal([]byte(input), &output)
}
