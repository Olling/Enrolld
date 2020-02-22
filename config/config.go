package config

import (
	"encoding/json"
	"os"

	l "github.com/Olling/Enrolld/logging"
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
		l.ErrorLog.Println("Error while reading the configuration file - Exiting")
		l.ErrorLog.Println(err)
		os.Exit(1)
	}

	_, err = os.Stat(Configuration.Path)

	if os.IsNotExist(err) {
		err := os.MkdirAll(Configuration.Path, 0744)
		if err != nil {
			l.ErrorLog.Println(err)
		} else {
			l.InfoLog.Println("Created: " + Configuration.Path)
		}
	}
}
