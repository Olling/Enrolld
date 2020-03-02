package fileio

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/config"
)


func DeleteServer(serverName string) error {
	return os.Remove(config.Configuration.Path + "/" + serverName)
}

func WriteToFile(server utils.ServerInfo, path string, append bool) (err error) {
	utils.SyncOutputMutex.Lock()
	defer utils.SyncOutputMutex.Unlock()

	bytes, marshalErr := json.MarshalIndent(server, "", "\t")
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
		slog.PrintError("ScriptPath is empty: \"" + config.Configuration.ScriptPath + "\"")
		return fmt.Errorf("ScriptPath is empty")
	} else {
		_, existsErr := os.Stat(config.Configuration.ScriptPath)

		if os.IsNotExist(existsErr) {
			slog.PrintError("ScriptPath does not exist: \"" + config.Configuration.ScriptPath + "\"")
			return fmt.Errorf("ScriptPath does not exist")
		}
	}
	return nil
}
