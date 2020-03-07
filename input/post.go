package input

import (
	"os"
	"fmt"
	"net"
	"os/exec"
	"io/ioutil"
	"time"
	"regexp"
	"errors"
	"strings"
	"syscall"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/output"
	"github.com/Olling/Enrolld/fileio"
	"github.com/Olling/Enrolld/metrics"
	"github.com/Olling/Enrolld/config"
)


func VerifyFQDN(serverid string, requestIP string) (string, error) {
	var fqdn string

	if serverid == "" {
		addresses, err := net.LookupAddr(requestIP)
		if err == nil && len(addresses) >= 1 {
			addr := addresses[0]
			if addr != "" {
				addr = strings.TrimSuffix(addr, ".")
				slog.PrintInfo("FQDN was empty (" + requestIP + ") but the IP had the following name: \"" + addr + "\"")
				serverid = addr
			}
		} else {
			return "", errors.New("FQDN is empty")
		}
	}

	if m, _ := regexp.MatchString("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$", serverid); !m {
		slog.PrintError("Received FQDN with illegal characters: \"", fqdn, "\" (", requestIP, ")")
		return "", errors.New("Received FQDN with illegal characters: \"" + fqdn + "\" (" + requestIP + ")")
	}

	if len(strings.Split(serverid, ".")) < 3 {
		addresses, err := net.LookupAddr(requestIP)
		if err == nil && len(addresses) >= 1 {
			addr := addresses[0]
			if addr != "" {
				addr = strings.TrimSuffix(addr, ".")
				slog.PrintInfo("Server \"" + serverid + "\"'s domain looks wrong - Replacing it with \"" + addr + "\"")
				serverid = addr
			}
		} else {
			slog.PrintError("Server \"" + serverid + "\"'s domain looks wrong, but no suitable name was found to replace it")
		}
	}
	return serverid, nil
}


func RemoveServer(serverID string) error {
	err := fileio.DeleteServer(config.Configuration.FileBackendDirectory + "/" + serverID)
	if err == nil {
		metrics.ServersDeleted.Inc()
	}
	return err
}

func UpdateServer(server utils.Server, isNewServer bool) error {
	server.LastSeen = time.Now().String()

	if !server.Exist() || isNewServer {
		isNewServer = true

		err := callEnrolldScript(server)
		if err != nil {
			slog.PrintError("Error running script against " + server.ServerID + "(" + server.IP + ")" + ": " + err.Error())
			utils.Notification("Enrolld failure", "Failed to enroll the following new server: " + server.ServerID + "(" + server.IP + ")", server)

			return err
		} else {
			slog.PrintInfo("Enrolld script successful: " + server.ServerID)
		}
	}

	var writeerr error
	writeerr = fileio.WriteToFile(server, config.Configuration.FileBackendDirectory + "/" + server.ServerID, false)

	if writeerr != nil {
		return writeerr
	} else {
		if isNewServer {
			slog.PrintInfo("Enrolled the following new machine:", server.ServerID, "(" + server.IP + ")")
			metrics.ServersAdded.Inc()
		} else {
			slog.PrintInfo("Updated the following machine:", server.ServerID, "(" + server.IP + ")")
			metrics.ServersUpdated.Inc()
		}
	}
	return nil
}


func callEnrolldScript(server utils.Server) (err error) {
	scriptPathErr := fileio.CheckScriptPath()

	if scriptPathErr != nil {
		return scriptPathErr
	}

	if server.ServerID == "" {
		slog.PrintError("Failed to call Enrollment Script - ServerID is empty!")
		return fmt.Errorf("ServerID was not given")
	}

	patchonly := false
	for _, inventory := range server.Groups {
		if inventory == "patchonly" {
			patchonly = true
		}
	}

	if !patchonly {
		tempDirectory := config.Configuration.TempPath + "/" + server.ServerID
		_, existsErr := os.Stat(tempDirectory)
		if os.IsNotExist(existsErr) {
			createErr := os.MkdirAll(tempDirectory, 0755)
			if createErr != nil {
				slog.PrintError(createErr)
				return fmt.Errorf("Could not create temp directory: " + tempDirectory)
			}
		}

		json, _ := output.GetInventoryInJSON([]utils.Server{server})
		json = strings.Replace(json, "\"", "\\\"", -1)

		ioutil.WriteFile(tempDirectory+"/singledynamicinventory", []byte("#!/bin/bash\necho \""+json+"\""), 0755)

		cmd := exec.Command("/bin/bash", config.Configuration.ScriptPath, tempDirectory+"/singledynamicinventory", server.ServerID)

		outfile, writeerr := os.Create(config.Configuration.LogPath + "/" + server.ServerID + ".log")
		if writeerr != nil {
			slog.PrintError("Error creating file", outfile.Name, writeerr)
		}

		defer outfile.Close()
		cmd.Stdout = outfile

		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		if startErr := cmd.Start(); err != nil {
			slog.PrintError("Could not start the enrolld script", startErr)
			return startErr
		}

		timer := time.AfterFunc(time.Duration(config.Configuration.Timeout) * time.Second, func() {
			slog.PrintError("The server " + server.ServerID + " have reached the timeout - Killing process", cmd.Process.Pid)
			pgid, err := syscall.Getpgid(cmd.Process.Pid)
			if err == nil {
				syscall.Kill(-pgid, 15)
			}
		})

		execErr := cmd.Wait()
		timer.Stop()

		if execErr != nil {
			slog.PrintError("Error while excecuting script. Please see the log for more info: " + config.Configuration.LogPath + "/" + server.ServerID + ".log")
			return execErr
		}
	}
	return nil
}
