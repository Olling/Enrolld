package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/Olling/Enrolld/api"
	"github.com/Olling/Enrolld/config"
	l "github.com/Olling/Enrolld/logging"
	"github.com/Olling/Enrolld/metrics"
	"github.com/Olling/Enrolld/output"
	"github.com/Olling/Enrolld/utils"
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

func checkScriptPath() (err error) {
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

func callEnrolldScript(server utils.ServerInfo) (err error) {
	scriptPathErr := checkScriptPath()

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

func httpHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/inventory":
			//httpGetInventory(w, r)
			fmt.Println("NOPE")
		case "/targets":
			//httpGetTargets(w, r)
			fmt.Println("NOPE")
		default:
			fmt.Fprintf(w, "running")
		}
	case "POST":
		httpPost(w, r)
	}
}

func httpPost(w http.ResponseWriter, r *http.Request) {
	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		var server utils.ServerInfo

		if r.FormValue("FQDN") != "" {
			server.FQDN = r.FormValue("FQDN")
		}

		if r.Header.Get("FQDN") != "" {
			server.FQDN = r.Header.Get("FQDN")
		}

		contentType := r.Header.Get("Content-type")
		switch contentType {
		case "application/json":
			if r.Body == nil {
				http.Error(w, "Please send a request body in JSON format", 400)
				return
			}

			err := json.NewDecoder(r.Body).Decode(&server)
			if err != nil {
				http.Error(w, "The received JSON body was in the wrong format", 400)
				return
			}
		default:
			server.Inventories = strings.Split(r.FormValue("inventory"), ",")

			if strings.TrimSpace(r.FormValue("AnsibleProperties")) != "" {
				if m, _ := regexp.MatchString("^([: ,_'\"a-zA-Z0-9]*)$", r.FormValue("AnsibleProperties")); !m {
					l.ErrorLog.Println("Received AnsibleProperties with illegal characters: \""+r.FormValue("AnsibleProperties")+"\" (", requestIP, ")")
				} else {
					jsonproperties := "{" + r.FormValue("AnsibleProperties") + " }"
					jsonproperties = strings.Replace(jsonproperties, "'", "\"", -1)

					var ansibleProperties map[string]string

					err := json.NewDecoder(strings.NewReader(jsonproperties)).Decode(&ansibleProperties)
					if err != nil {
						http.Error(w, "The AnsibleProperties was in the wrong format", 400)
						return
					}

					server.AnsibleProperties = ansibleProperties
				}
			}
		}

		isNewServer := false

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
				l.ErrorLog.Println("FQDN is empty (", requestIP, ")")
				http.Error(w, http.StatusText(500), 500)
				return
			}
		}

		if m, _ := regexp.MatchString("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$", server.FQDN); !m {
			l.ErrorLog.Println("Received FQDN with illegal characters: \"", server.FQDN, "\" (", requestIP, ")")
			http.Error(w, http.StatusText(500), 500)
			return
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

		server.IP = requestIP
		server.LastSeen = time.Now().String()

		for _, fqdn := range config.Configuration.Blacklist {
			if strings.ToLower(server.FQDN) == strings.ToLower(fqdn) {
				l.InfoLog.Println(server.FQDN + " (" + server.IP + ") is on the blacklist - Ignoring")
				fmt.Fprintln(w, "Ignored")
				return
			}
		}

		if strings.ToLower(r.FormValue("NewServer")) == "true" || strings.ToLower(r.FormValue("NewServer")) == "yes" || strings.ToLower(server.NewServer) == "true" || strings.ToLower(server.NewServer) == "yes" {
			isNewServer = true
		}

		if !server.Exist() || isNewServer {
			isNewServer = true

			enrolldErr := callEnrolldScript(server)
			if enrolldErr != nil {
				l.ErrorLog.Println("Error running script against " + server.FQDN + "(" + server.IP + ")" + ": " + enrolldErr.Error())

				if val, ok := server.AnsibleProperties["global_server_type"]; !ok {
					notification("Enrolld failure", "Failed to enroll the following new server: "+server.FQDN+"("+server.IP+")", server)
				} else {
					if val != "clud" {
						notification("Enrolld failure", "Failed to enroll the following new server: "+server.FQDN+"("+server.IP+")", server)
					}
				}

				http.Error(w, http.StatusText(500), 500)
				return
			} else {
				l.InfoLog.Println("Enrolld script successful: " + server.FQDN)
			}
		}

		var writeerr error
		writeerr = WriteToFile(server, config.Configuration.Path+"/"+server.FQDN, false)

		if writeerr != nil {
			l.ErrorLog.Println("Write Error")
			http.Error(w, http.StatusText(500), 500)
		} else {
			fmt.Fprintln(w, "Enrolled")

			if isNewServer {
				l.InfoLog.Println("Enrolled the following new machine: " + server.FQDN + " (" + server.IP + ")")
			} else {
				l.InfoLog.Println("Updated the following machine: " + server.FQDN + " (" + server.IP + ")")
			}
		}
	}
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

func main() {
	l.InitializeLogging(os.Stdout, os.Stderr)
	config.InitializeConfiguration("/etc/enrolld/enrolld.conf")

	scriptPathErr := checkScriptPath()
	if scriptPathErr != nil {
		log.Fatal("ScriptPath Problem - stopping")
	}

	metrics.Init()
	api.SetupRouter()
	// http.HandleFunc("/", httpHandler)

	// go http.ListenAndServe(":"+configuration.Port, nil)
	// tlserr := http.ListenAndServeTLS(":"+configuration.TlsPort, configuration.TlsCert, configuration.TlsKey, nil)
	// if tlserr != nil {
	// 	errorlog.Println("Error starting TLS: ", tlserr)
	// }
	// errorlog.Println("Error happend while serving port")
}
