package io

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
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/output"
	"github.com/Olling/Enrolld/fileio"
	"github.com/Olling/Enrolld/config"
	l "github.com/Olling/Enrolld/logging"
)

func VerifyFQDN(server utils.ServerInfo, requestIP string) (string, error) {
	var fqdn string

	if server.FQDN == "" {
		addresses, err := net.LookupAddr(requestIP)
		if err == nil && len(addresses) >= 1 {
			addr := addresses[0]
			if addr != "" {
				addr = strings.TrimSuffix(addr, ".")
				l.InfoLog.Println("FQDN was empty (" + requestIP + ") but the IP had the following name: \"" + addr + "\"")
				server.FQDN = addr
			}
		} else {
			return "", errors.New("FQDN is empty")
		}
	}

	if m, _ := regexp.MatchString("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$", server.FQDN); !m {
		l.ErrorLog.Println("Received FQDN with illegal characters: \"", fqdn, "\" (", requestIP, ")")
		return "", errors.New("Received FQDN with illegal characters: \"" + fqdn + "\" (" + requestIP + ")")
	}

	if len(strings.Split(server.FQDN, ".")) < 3 {
		addresses, err := net.LookupAddr(requestIP)
		if err == nil && len(addresses) >= 1 {
			addr := addresses[0]
			if addr != "" {
				addr = strings.TrimSuffix(addr, ".")
				l.InfoLog.Println("Server \"" + server.FQDN + "\"'s domain looks wrong - Replacing it with \"" + addr + "\"")
				server.FQDN = addr
			}
		} else {
			l.ErrorLog.Println("Server \"" + server.FQDN + "\"'s domain looks wrong, but no suitable name was found to replace it")
		}
	}
	return server.FQDN, nil
}

//func GetContent () {
//
//
//	default:
//		server.Inventories = strings.Split(r.FormValue("inventory"), ",")
//
//		if strings.TrimSpace(r.FormValue("AnsibleProperties")) != "" {
//			if m, _ := regexp.MatchString("^([: ,_'\"a-zA-Z0-9]*)$", r.FormValue("AnsibleProperties")); !m {
//				l.ErrorLog.Println("Received AnsibleProperties with illegal characters: \""+r.FormValue("AnsibleProperties")+"\" (", requestIP, ")")
//			} else {
//				jsonproperties := "{" + r.FormValue("AnsibleProperties") + " }"
//				jsonproperties = strings.Replace(jsonproperties, "'", "\"", -1)
//
//				var ansibleProperties map[string]string
//
//				err := json.NewDecoder(strings.NewReader(jsonproperties)).Decode(&ansibleProperties)
//				if err != nil {
//					http.Error(w, "The AnsibleProperties was in the wrong format", 400)
//					return
//				}
//
//				server.AnsibleProperties = ansibleProperties
//			}
//		}
//	}
//}



func UpdateServer(server utils.ServerInfo, isNewServer bool) error {
	server.LastSeen = time.Now().String()

	if !server.Exist() || isNewServer {
		isNewServer = true

		enrolldErr := callEnrolldScript(server)
		if enrolldErr != nil {
			l.ErrorLog.Println("Error running script against " + server.FQDN + "(" + server.IP + ")" + ": " + enrolldErr.Error())

			if val, ok := server.AnsibleProperties["global_server_type"]; !ok {
				notification("Enrolld failure", "Failed to enroll the following new server: " + server.FQDN + "(" + server.IP + ")", server)
			} else {
				if val != "clud" {
					notification("Enrolld failure", "Failed to enroll the following new server: " + server.FQDN + "(" + server.IP + ")", server)
				}
			}

			return enrolldErr
		} else {
			l.InfoLog.Println("Enrolld script successful: " + server.FQDN)
		}
	}

	var writeerr error
	writeerr = fileio.WriteToFile(server, config.Configuration.Path + "/" + server.FQDN, false)

	if writeerr != nil {
		return writeerr
	} else {
		if isNewServer {
			l.InfoLog.Println("Enrolled the following new machine: " + server.FQDN + " (" + server.IP + ")")
		} else {
			l.InfoLog.Println("Updated the following machine: " + server.FQDN + " (" + server.IP + ")")
		}
	}
	return nil
}


func notification(subject string, message string, server utils.ServerInfo) {
	binary, err := exec.LookPath(config.Configuration.NotificationScriptPath)
	if err != nil {
		l.ErrorLog.Println("Could not find the notification script in the given path", config.Configuration.NotificationScriptPath, err)
	}
	cmd := exec.Command(binary)

	env := os.Environ()
	env = append(env, fmt.Sprintf("SUBJECT=%s", subject))
	env = append(env, fmt.Sprintf("MESSAGE=%s", message))

	env = append(env, fmt.Sprintf("SERVER_FQDN=%s", server.FQDN))
	env = append(env, fmt.Sprintf("SERVER_IP=%s", server.IP))
	env = append(env, fmt.Sprintf("SERVER_PROPERTIES=%s", server.AnsibleProperties))
	env = append(env, fmt.Sprintf("SERVER_INVENTORIES=%s", server.Inventories))
	env = append(env, fmt.Sprintf("SERVER_LASTSEEN=%s", server.LastSeen))

	cmd.Env = env

	startErr := cmd.Start()
	if startErr != nil {
		l.ErrorLog.Println("Could not send notification", startErr)
	}

	cmd.Wait()
}


func callEnrolldScript(server utils.ServerInfo) (err error) {
	scriptPathErr := fileio.CheckScriptPath()

	if scriptPathErr != nil {
		return scriptPathErr
	}

	if server.FQDN == "" {
		l.ErrorLog.Println("FQDN is empty!")
		return fmt.Errorf("FQDN was not given")
	}

	patchonly := false
	for _, inventory := range server.Inventories {
		if inventory == "patchonly" {
			patchonly = true
		}
	}

	if !patchonly {
		tempDirectory := config.Configuration.TempPath + "/" + server.FQDN
		_, existsErr := os.Stat(tempDirectory)
		if os.IsNotExist(existsErr) {
			createErr := os.MkdirAll(tempDirectory, 0755)
			if createErr != nil {
				l.ErrorLog.Println(createErr)
				return fmt.Errorf("Could not create temp directory: " + tempDirectory)
			}
		}

		json, _ := output.GetInventoryInJSON([]utils.ServerInfo{server})
		json = strings.Replace(json, "\"", "\\\"", -1)

		ioutil.WriteFile(tempDirectory+"/singledynamicinventory", []byte("#!/bin/bash\necho \""+json+"\""), 0755)

		//TODO Fix This
		cmd := exec.Command("sudo", "-u", "USER", "/bin/bash", config.Configuration.ScriptPath, tempDirectory+"/singledynamicinventory", server.FQDN)

		outfile, writeerr := os.Create(config.Configuration.LogPath + "/" + server.FQDN + ".log")
		if writeerr != nil {
			l.ErrorLog.Println("Error creating file", outfile.Name, writeerr)
		}

		defer outfile.Close()
		cmd.Stdout = outfile

		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		if startErr := cmd.Start(); err != nil {
			l.ErrorLog.Println("Could not start the enrolld script", startErr)
			return startErr
		}

		timer := time.AfterFunc(time.Duration(config.Configuration.Timeout)*time.Second, func() {
			l.ErrorLog.Println("The server "+server.FQDN+" have reached the timeout - Killing process", cmd.Process.Pid)
			pgid, err := syscall.Getpgid(cmd.Process.Pid)
			if err == nil {
				syscall.Kill(-pgid, 15)
			}
		})

		execErr := cmd.Wait()
		timer.Stop()

		if execErr != nil {
			l.ErrorLog.Println("Error while excecuting script. Please see the log for more info: " + config.Configuration.LogPath + "/" + server.FQDN + ".log")
			return execErr
		}
	}
	return nil
}
