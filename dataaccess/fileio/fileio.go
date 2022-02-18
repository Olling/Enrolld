package fileio

import (
	"os"
	"fmt"
	"path"
	"sync"
	"time"
	"errors"
	"regexp"
	"strings"
	"syscall"
	"strconv"
	"os/exec"
	"io/ioutil"
	"math/rand"
	"encoding/json"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/metrics"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enrolld/dataaccess/config"
)

var (
	SyncOutputMutex		sync.Mutex
)

func DeleteServer(serverName string) error {
	return os.Remove(config.Configuration.FileBackendDirectory + "/" + serverName)
}

func WriteToFile(filepath string, content string, appendToFile bool, filemode os.FileMode) (err error) {
	SyncOutputMutex.Lock()
	defer SyncOutputMutex.Unlock()

	if appendToFile {
		file, fileerr := os.OpenFile(filepath, os.O_APPEND, filemode)
		defer file.Close()
		if fileerr != nil {
			return fileerr
		}

		_, writeerr := file.WriteString(content)
		return writeerr
	} else {
		err := ioutil.WriteFile(filepath, []byte(content), filemode)
		if err != nil {
			slog.PrintError("Error while writing file", err)
			return err
		}
		return nil
	}
}

func WriteStructToFile(s interface{}, filepath string, appendToFile bool) (err error) {
	json, err := utils.StructToJson(s)

	if err != nil {
		slog.PrintError("Could not convert struct to json", err)
		return err
	}

	return WriteToFile(filepath, json, appendToFile, 0664)

}

func CheckScriptPath(filepath string) error {
	if filepath == "" {
		slog.PrintError("ScriptPath is empty")
		return fmt.Errorf("ScriptPath is empty")
	}

	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		slog.PrintError("ScriptPath does not exist: '" + filepath + "'")
		return fmt.Errorf("ScriptPath does not exist")
	}

	return nil
}

func LoadFromFile(s interface{}, filepath string) error {
	file, err := os.Open(filepath)
	defer file.Close()

	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&s)
	file.Close()

	return nil
}

func LoadOverwrites(overwrites interface{}) error {
	return LoadFromFile(&overwrites, config.Configuration.FileBackendDirectory + "/overwrites.json")
}

func GetFileList(directoryPath string) ([]os.FileInfo, error) {
	filelist, err := ioutil.ReadDir(directoryPath)
	return filelist, err
}

func LoadScripts(scripts map[string]objects.Script) error {
	scripts = make(map[string]objects.Script)

	filelist, err := GetFileList(config.Configuration.ScriptDirectory)
	if err != nil {
		slog.PrintDebug("Failed to load script list", err)
		return err
	}

	for _,directory := range filelist {
		if !directory.IsDir() {
			slog.PrintDebug("Ignoring the following file in the script path", directory.Name())
			continue
		}

		var script objects.Script
		scriptID := directory.Name()
		scriptPath := path.Join(config.Configuration.ScriptDirectory, scriptID)

		err = LoadFromFile(&script, path.Join(scriptPath, scriptID + ".json"))
		if err != nil{
			slog.PrintError("Failed to get script information from", path.Join(scriptPath, scriptID + ".json"))
			continue
		}

		//TODO there might be a reference problem here
		scripts[scriptID] = script
	}

	return nil
}


func SaveOverwrites(overwrites interface{}) error {
	return WriteStructToFile(overwrites, config.Configuration.FileBackendDirectory + "/overwrites.json", false)
}

func FileExist(filepath string) bool {
	_, existsErr := os.Stat(filepath)

	if os.IsNotExist(existsErr) {
		return false
	} else {
		return true
	}
}

func AddOverwrites(server *objects.Server, overwrites map[string]objects.Overwrite) {
	for _,overwrite := range overwrites {
		matchServerID,_ := regexp.MatchString(overwrite.ServerIDRegexp, server.ServerID)

		matchAnsibleInventories := false
		for _,inventory := range server.Groups {
			if match,_ := regexp.MatchString(overwrite.GroupRegexp, inventory); match {
				matchAnsibleInventories = true
				break
			}
		}

		matchAnsibleProperties,_ := regexp.MatchString(overwrite.PropertiesRegexp.Value, server.Properties[overwrite.PropertiesRegexp.Key])

		if matchServerID && matchAnsibleInventories && matchAnsibleProperties {
			server.Groups = append(server.Groups, overwrite.Groups...)


			for key,value := range overwrite.Properties {
				server.Properties[key] = value
			}
		}
	}
}

func RunScript(scriptPath string, server objects.Server, scriptID string, timeout int) error {
	err := CheckScriptPath(scriptPath)
	if err != nil {
		return err
	}

	tempDirectory := path.Join(config.Configuration.TempPath, server.ServerID, strconv.Itoa(rand.Intn(200)))
	inventoryPath := path.Join(tempDirectory, "single.inventory")

	logPath := path.Join(config.Configuration.LogPath, scriptID, server.ServerID + ".log")

	_, err = os.Stat(path.Join(config.Configuration.LogPath, scriptID))
	if os.IsNotExist(err) {
		err := os.MkdirAll(path.Join(config.Configuration.LogPath, scriptID), 0744)
		if err != nil {
			slog.PrintDebug("Failed to create log directory", err)
		}
	}

	outfile, err := os.Create(logPath)
	if err != nil {
		slog.PrintError("Error creating logfile", outfile.Name, err)
	}
	defer outfile.Close()

	_, existsErr := os.Stat(tempDirectory)
	if os.IsNotExist(existsErr) {
		createErr := os.MkdirAll(tempDirectory, 0755)
		if createErr != nil {
			slog.PrintError(createErr)
			return fmt.Errorf("Could not create temp directory: " + tempDirectory)
		}
	}

	json, _ := utils.GetInventoryInJSON([]objects.Server{server})
	json = strings.Replace(json, "\"", "\\\"", -1)
	inventory := "#!/bin/bash\necho \"" + json + "\""

	WriteToFile(inventoryPath, inventory, false, 0755)

	cmd := exec.Command("/bin/bash", scriptPath, inventoryPath, server.ServerID)
	cmd.Stdout = outfile
	cmd.Stderr = outfile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err = cmd.Start(); err != nil {
		slog.PrintError("Could not start the script", scriptID, err)
		return err
	}

	timer := time.AfterFunc(time.Duration(timeout) * time.Second, func() {
		slog.PrintError("The script ", scriptID + "(" + server.ServerID + ")", "has reached the timeout - Killing process", cmd.Process.Pid)
		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, 15)
		}
	})

	execErr := cmd.Wait()
	timer.Stop()

	if execErr != nil {
		slog.PrintError("Error while excecuting script", scriptID, "Please see the log for more info:", logPath)
		return execErr
	}

	return nil
}

func GetServers(overwrites map[string]objects.Overwrite) ([]objects.Server, error) {
	var inventory []objects.Server

	filelist, err := GetFileList(config.Configuration.FileBackendDirectory)
	if err != nil {
		slog.PrintError("Failed to get inventory:", err)
		return nil, err
	}

//	utils.SyncGetInventoryMutex.Lock()
//	defer utils.SyncGetInventoryMutex.Unlock()

	for _, child := range filelist {
		if child.IsDir() == false {
			server, err := GetServer(child.Name(), overwrites)

			if err != nil {
				slog.PrintDebug("Could not get server:", config.Configuration.FileBackendDirectory + "/" + child.Name(), "Reason:", err)
				continue
			}

			inventory = append(inventory, server)
		}
	}
	return inventory, nil
}

func ServerExist(serverID string) bool {
	return FileExist(config.Configuration.FileBackendDirectory + "/" + serverID)
}

func RemoveServer(serverID string) error {
	err := DeleteServer(config.Configuration.FileBackendDirectory + "/" + serverID)
	if err == nil {
		metrics.ServersDeleted.Inc()
	}
	return err
}

func GetServer(serverID string, overwrites map[string]objects.Overwrite) (server objects.Server, err error) {
	err = LoadFromFile(&server, config.Configuration.FileBackendDirectory + "/" + serverID)

	if err != nil {
		return server, err
	}

	AddOverwrites(&server, overwrites)

	layout := "2006-01-02 15:04:05.999999999 -0700 MST"

	if strings.Contains(server.LastSeen, "m=") {
		server.LastSeen = strings.Split(server.LastSeen, " m=")[0]
	}

	date, err := time.Parse(layout, server.LastSeen)
	if err != nil {
		return server, err
	}

	date = date.Add(time.Minute * time.Duration(config.Configuration.MaxAgeInMinutes))
	if date.After(time.Now()) {
		return server,nil
	}

	return server, errors.New("Server was beyond max age")
}


func UpdateServer(server objects.Server) (writeerr error) {
	server.LastSeen = time.Now().String()

	writeerr = WriteStructToFile(server, config.Configuration.FileBackendDirectory + "/" + server.ServerID, false)

	if writeerr == nil {
		slog.PrintInfo("Updated the following machine:", server.ServerID, "(" + server.IP + ")")
		metrics.ServersUpdated.Inc()
	}

	return writeerr
}
