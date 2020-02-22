package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type configuration struct {
	Path                   string
	ScriptPath             string
	NotificationScriptPath string
	TempPath               string
	LogPath                string
	Port                   string
	MaxAgeInDays           int
	DefaultInventoryName   string

	TargetsPort string
	TlsPort     string
	TlsCert     string
	TlsKey      string
	Blacklist   []string
	Timeout     int
}

var Configuration configuration

func InitializeConfiguration(path string) {
	file, _ := os.Open(path)

	err := json.NewDecoder(file).Decode(&Configuration)
	if err != nil {
		// needs looging (errorlog)
		fmt.Println("Error while reading the configuration file - Exiting")
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = os.Stat(Configuration.Path)

	if os.IsNotExist(err) {
		err := os.MkdirAll(Configuration.Path, 0744)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Created: " + Configuration.Path)
		}
	}
}
