package io

import (
	"encoding/json"
	"os"
	"fmt"
	"io/ioutil"
	"github.com/Olling/Enrolld/utils"
	l "github.com/Olling/Enrolld/logging"
	"github.com/Olling/Enrolld/config"
)

func WriteToFile(server utils.ServerInfo, path string, append bool) (err error) {
	utils.SyncOutputMutex.Lock()
	defer utils.SyncOutputMutex.Unlock()

	server.NewServer = ""
	bytes, marshalErr := json.MarshalIndent(server, "", "\t")
	if marshalErr != nil {
		l.ErrorLog.Println("Error while converting to json")
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
			l.ErrorLog.Println("Error while writing file")
			l.ErrorLog.Println(err)
			return err
		}
		return nil
	}
}

func CheckScriptPath() (err error) {
	if config.Configuration.ScriptPath == "" {
		l.ErrorLog.Println("ScriptPath is empty: \"" + config.Configuration.ScriptPath + "\"")
		return fmt.Errorf("ScriptPath is empty")
	} else {
		_, existsErr := os.Stat(config.Configuration.ScriptPath)

		if os.IsNotExist(existsErr) {
			l.ErrorLog.Println("ScriptPath does not exist: \"" + config.Configuration.ScriptPath + "\"")
			return fmt.Errorf("ScriptPath does not exist")
		}
	}
	return nil
}
