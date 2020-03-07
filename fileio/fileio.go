package fileio

import (
	"os"
	"fmt"
	"regexp"
	"io/ioutil"
	"encoding/json"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/config"
)


func DeleteServer(serverName string) error {
	return os.Remove(config.Configuration.FileBackendDirectory + "/" + serverName)
}

func WriteToFile(s interface{}, path string, append bool) (err error) {
	utils.SyncOutputMutex.Lock()
	defer utils.SyncOutputMutex.Unlock()

	bytes, marshalErr := json.MarshalIndent(s, "", "\t")
	if marshalErr != nil {
		slog.PrintError("Error while converting to json")
		return marshalErr
	}
	content := string(bytes)

	if append {
		file, fileerr := os.OpenFile(path, os.O_APPEND, 644)
		defer file.Close()
		if fileerr != nil {
			return fileerr
		}

		_, writeerr := file.WriteString(content)
		return writeerr
	} else {
		err := ioutil.WriteFile(path, []byte(content), 0644)
		if err != nil {
			slog.PrintError("Error while writing file")
			slog.PrintError(err)
			return err
		}
		return nil
	}
}

func CheckScriptPath() (err error) {
	if config.Configuration.ScriptPath == "" {
		slog.PrintError("ScriptPath is empty: '" + config.Configuration.ScriptPath + "'")
		return fmt.Errorf("ScriptPath is empty")
	} else {
		_, existsErr := os.Stat(config.Configuration.ScriptPath)

		if os.IsNotExist(existsErr) {
			slog.PrintError("ScriptPath does not exist: '" + config.Configuration.ScriptPath + "'")
			return fmt.Errorf("ScriptPath does not exist")
		}
	}
	return nil
}

func LoadFromFile(s interface{}, path string) error {
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&s)
	file.Close()

	return nil
}

func LoadOverwrites() {
	LoadFromFile(&utils.Overwrites, config.Configuration.FileBackendDirectory + "/overwrites.json")
}

func SaveOverwrites() {
	err := WriteToFile(utils.Overwrites, config.Configuration.FileBackendDirectory + "/overwrites.json", false)
	if err != nil {
		slog.PrintError("Failed to write AnsibleAddons:", err)
	}
}

func AddOverwrites(server *utils.Server) {
	for _,overwrite := range utils.Overwrites {
		matchServerID,_ := regexp.MatchString(overwrite.ServerIDRegexp, server.ServerID)

		matchAnsibleInventories := false
		for _,inventory := range server.Groups {
			if match,_ := regexp.MatchString(overwrite.InventoriesRegexp, inventory); match {
				matchAnsibleInventories = true
				break
			}
		}

		matchAnsibleProperties,_ := regexp.MatchString(overwrite.PropertiesRegexp.Value, server.Properties[overwrite.PropertiesRegexp.Key])

		if matchServerID && matchAnsibleInventories && matchAnsibleProperties {
			server.Groups = append(server.Groups, overwrite.Inventories...)


			for key,value := range overwrite.Properties {
				server.Properties[key] = value
			}
		}
	}
}
