package input

import (
	"fmt"
	"net"
	"time"
	"regexp"
	"errors"
	"strings"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
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
				slog.PrintDebug("Server \"" + serverid + "\"'s domain looks wrong - Replacing it with \"" + addr + "\"")
				serverid = addr
			}
		} else {
			slog.PrintError("Server \"" + serverid + "\"'s domain looks wrong, but no suitable name was found to replace it")
		}
	}
	return serverid, nil
}

func ServerExist(server utils.Server) bool {
	return fileio.FileExist(config.Configuration.FileBackendDirectory + "/" + server.ServerID)
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

	if !ServerExist(server) || isNewServer {
		isNewServer = true

		err := RunScript(config.Configuration.EnrollmentScriptPath,server, "Enroll", config.Configuration.Timeout)
		if err != nil {
			slog.PrintError("Error running script against", server.ServerID, "(" + server.IP + "):", err)
			utils.Notification("Enrolld failure", "Failed to enroll the following new server: " + server.ServerID + "(" + server.IP + ")", server)

			return err
		} else {
			slog.PrintInfo("Enrolld script successful: " + server.ServerID)
		}
	}

	var writeerr error
	writeerr = fileio.WriteStructToFile(server, config.Configuration.FileBackendDirectory + "/" + server.ServerID, false)

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


func RunScript(scriptPath string, server utils.Server, scriptID string, timeout int) error {
	if server.ServerID == "" {
		slog.PrintError("Failed to call", scriptID, "script - ServerID is empty!")
		return fmt.Errorf("ServerID was not given")
	}

	err := fileio.RunScript(scriptPath, server, scriptID, timeout)

	return err
}
