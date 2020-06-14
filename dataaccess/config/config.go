package config

import (
	"os"
	"encoding/json"
	"github.com/Olling/slog"
)


type configuration struct {
	FileBackendDirectory	string
	ScriptDirectory		string
	EnrollmentScriptPath	string
	NotificationScriptPath	string
	TempPath		string
	LogPath			string
	Port			string
	MaxAgeInMinutes		int
	DefaultInventoryName	string

	TargetsPort string
	TlsPort     string
	TlsCert     string
	TlsKey      string
	Blacklist   []string
	Timeout     int
}


var (
	Configuration	configuration
)


func InitializeConfiguration(path string) {
	file, _ := os.Open(path)

	err := json.NewDecoder(file).Decode(&Configuration)
	if err != nil {
		slog.PrintError("Error while reading the configuration file - Exiting")
		slog.PrintError(err)
		os.Exit(1)
	}

	_, err = os.Stat(Configuration.FileBackendDirectory)

	if os.IsNotExist(err) {
		err := os.MkdirAll(Configuration.FileBackendDirectory, 0744)
		if err != nil {
			slog.PrintError(err)
		} else {
			slog.PrintInfo("Created: " + Configuration.FileBackendDirectory)
		}
	}

}
