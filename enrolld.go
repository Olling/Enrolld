package main

import (
    	"fmt"
	"sort"
    	"net/http"
	"net"
	"sync"
	"io"
	"io/ioutil"
	"encoding/json"
	"crypto/sha1"
	"os"
	"log"
	"strings"
	"time"
	"os/exec"
	"regexp"
	"syscall"
)

type ServerInfo struct {
	FQDN string
	IP string
        LastSeen string
	NewServer string `json:"NewServer,omitempty"`
	Inventories []string
	AnsibleProperties map[string]string
}

type TargetList struct {
	Targets []string `json:"targets"`
	Labels map[string]string `json:"labels"`
}

func (server ServerInfo) Exist () bool {
   	_, existsErr := os.Stat(configuration.Path + "/" + server.FQDN)

        if os.IsNotExist(existsErr) {
        	return false
	} else {
		return true
	}
}

type Configuration struct {
	Path string
	ScriptPath string
	NotificationScriptPath string
	TempPath string
	LogPath string
	Port string
	MaxAgeInDays int

	TargetsPort string
	TlsPort string
	TlsCert string
	TlsKey string
	Blacklist []string
	Timeout int
}

var (
	infolog *log.Logger
	errorlog *log.Logger
 	syncOutputMutex sync.Mutex
 	syncGetInventoryMutex sync.Mutex
 	configuration Configuration
)


func CategorizeInventories(inventories []ServerInfo) ([]string,map[string][]ServerInfo) {
        keys := make([]string,0)
	results := make(map[string][]ServerInfo)

        for _, inventory := range inventories {
                for _, foundInventoryName := range inventory.Inventories {
                        if results[foundInventoryName] != nil {
				results[foundInventoryName] = append(results[foundInventoryName], inventory)
			} else {
				keys = append(keys,foundInventoryName)
				results[foundInventoryName] = []ServerInfo{inventory}
                        }
                }
        }

        return keys,results
}


func GetInventoryInJSON(inventories []ServerInfo) (string,error) {
	inventoryjson := "{"	

	keys,inventoryMap := CategorizeInventories(inventories) 

        inventoryjson += "\n\t\"" + configuration.DefaultInventoryName + "\"\t: {\n\t\"hosts\"\t: ["
	for _,inventory := range inventories {
		inventoryjson += "\"" + inventory.FQDN + "\", "
	}
 	inventoryjson = strings.TrimSuffix(inventoryjson,", ")
      	inventoryjson += "]\n\t},"

	for _,key := range keys {
		inventoryjson += "\n\t\"" + key + "\"\t: {\n\t\"hosts\"\t: ["
		for _,inventory := range inventoryMap[key] {
			inventoryjson += "\"" + inventory.FQDN + "\", "
		}
		inventoryjson = strings.TrimSuffix(inventoryjson,", ")
        	inventoryjson += "]\n\t},"
	}

	inventoryjson += "\n\t\"_meta\" : {\n\t\t\"hostvars\" : {"

	for _,server := range inventories {
		if len(server.AnsibleProperties) != 0 {
			propertiesjsonbytes, err := json.Marshal(server.AnsibleProperties)
			if err != nil {
				errorlog.Println("Error in converting map to json",err)
			} else {
				propertiesjson := string(propertiesjsonbytes)
				propertiesjson = strings.TrimPrefix(propertiesjson, "{")
				propertiesjson = strings.TrimSuffix(propertiesjson, "}")
		        	inventoryjson += "\n\t\t\t\"" + server.FQDN + "\": {\n\t\t\t\t" + propertiesjson + "\n\t\t\t},"
			}
		}
	}

        inventoryjson = strings.TrimSuffix(inventoryjson,",")
	inventoryjson += "\n\t\t}\n\t}\n}"
	
	return inventoryjson,nil
}


func GetTargetsInJSON(servers []ServerInfo) (string,error) {
	entriesmap := make(map[string]TargetList)

	for _,server := range servers {
		if configuration.TargetsPort != "" {
			server.FQDN = server.FQDN + ":" + configuration.TargetsPort
		}

		var entry TargetList
		entry.Targets = []string{server.FQDN}
		if server.AnsibleProperties != nil {
			entry.Labels = server.AnsibleProperties
		} else {
			entry.Labels = make(map[string]string)
		}

		inventories := strings.Join(server.Inventories,", ")
		entry.Labels["inventories"] = inventories

		var label string
		if len(entry.Labels) == 0 {
			label = "nolabels"
		} else {
			sha1calc := sha1.New()
			
			var keys []string
			for key,_ := range entry.Labels {
				keys = append(keys,key)
			}
			sort.Strings(keys)

			for _,key := range keys {
				io.WriteString(sha1calc, key + ":" + entry.Labels[key])
			}
			label = fmt.Sprintf("%x",sha1calc.Sum(nil))
		}

		_, keyexists := entriesmap[label]
		if keyexists {
			tempentry := entriesmap[label]
			tempentry.Targets = append(tempentry.Targets, entry.Targets...)
			entriesmap[label] = tempentry
		} else {
			entriesmap[label] = entry
		}
	}

	var entries []TargetList
	for _,value := range entriesmap {
		entries = append(entries,value)
	}
	entriesjson,err := ToJson(entries)

	return entriesjson,err
}


func GetInventory (path string) ([]ServerInfo, error) {
	var inventories []ServerInfo
       
	filelist, filelisterr := ioutil.ReadDir(path)
        if filelisterr != nil {
                errorlog.Println(filelisterr)
		return nil, filelisterr
        }

 	syncGetInventoryMutex.Lock()
    	defer syncGetInventoryMutex.Unlock()

        for _, child := range filelist {
		if child.IsDir() == false {
	   		file,fileerr := os.Open(path + "/" +child.Name())
			
			if fileerr != nil {
                                errorlog.Println("Error while reading file",path + "/" + child.Name(),"Reason:",fileerr)
				continue
			}

        		decoder := json.NewDecoder(file)
			var inventory ServerInfo
        		err := decoder.Decode(&inventory)

        		if err != nil {
                		errorlog.Println("Error while decoding file",path + "/" + child.Name(),"Reason:",err)
        		} else {
				layout := "2006-01-02 15:04:05.999999999 -0700 MST"

				if strings.Contains(inventory.LastSeen,"m=") {
					inventory.LastSeen = strings.Split(inventory.LastSeen," m=")[0]
				}

				date,parseErr := time.Parse(layout,inventory.LastSeen)

				if parseErr != nil {
					errorlog.Println("Could not parse date")
					errorlog.Println(parseErr)
				}

				date = date.AddDate(0,0,configuration.MaxAgeInDays)

				if date.After(time.Now()) {
					inventories = append(inventories,inventory)
				}
			}
		}
        }
	return inventories, nil
}

func ToJson(s interface{}) (string, error) {
        bytes, marshalErr := json.MarshalIndent(s,"","\t")
	return string(bytes), marshalErr
}

func FromJson(input string,output interface{}) (error) {
	return json.Unmarshal([]byte(input), &output)
}


func WriteToFile(server ServerInfo, path string, append bool) (err error){
    	syncOutputMutex.Lock()
    	defer syncOutputMutex.Unlock()

	server.NewServer = ""	
    	bytes, marshalErr := json.MarshalIndent(server,"","\t")
    	if marshalErr != nil {
		errorlog.Println("Error while converting to json")
        	return marshalErr 
    	}
	content := string(bytes)

	if append {
		file, fileerr := os.OpenFile(path, os.O_APPEND, 644)
		defer file.Close()
		if fileerr != nil {
			return fileerr
		}

		_,writeerr := file.WriteString(content)
		return writeerr
	} else {
		err := ioutil.WriteFile(path, []byte(content), 0644)
	    	if err != nil {
                	errorlog.Println("Error while writing file")
			errorlog.Println(err)
			return err
		}
		return nil
    	}
}

func checkScriptPath () (err error){
       if configuration.ScriptPath == "" {
                errorlog.Println("ScriptPath is empty: \"" + configuration.ScriptPath + "\"")
                return fmt.Errorf("ScriptPath is empty")
        } else {
                _, existsErr := os.Stat(configuration.ScriptPath)

                if os.IsNotExist(existsErr) {
                        errorlog.Println("ScriptPath does not exist: \"" + configuration.ScriptPath + "\"")
                        return fmt.Errorf("ScriptPath does not exist")
                }
        }
	return nil
}


func callEnrolldScript(server ServerInfo) (err error) {
	scriptPathErr := checkScriptPath()

	if scriptPathErr != nil {
	        return scriptPathErr
	}

   	if server.FQDN == "" {
                errorlog.Println("FQDN is empty!")
                return fmt.Errorf("FQDN was not given")
        }

	patchonly := false
        for _, inventory := range server.Inventories {
		if inventory == "patchonly" {
			patchonly = true
		}
	}

	if !patchonly {
		tempDirectory := configuration.TempPath + "/" + server.FQDN
		_, existsErr := os.Stat(tempDirectory)
		if os.IsNotExist(existsErr) {
			createErr := os.MkdirAll(tempDirectory,0755)
			if createErr != nil {
				errorlog.Println(createErr)
				return fmt.Errorf("Could not create temp directory: " + tempDirectory)
			}
		}

		json,_ := GetInventoryInJSON([]ServerInfo{server})
		json = strings.Replace(json, "\"", "\\\"", -1)

		ioutil.WriteFile(tempDirectory + "/singledynamicinventory", []byte("#!/bin/bash\necho \"" + json + "\""), 0755)

		//TODO Fix This
		cmd := exec.Command("sudo","-u","USER","/bin/bash", configuration.ScriptPath, tempDirectory + "/singledynamicinventory",server.FQDN)
	
          	outfile, writeerr := os.Create(configuration.LogPath + "/" + server.FQDN + ".log")
                if writeerr != nil {
                        errorlog.Println("Error creating file",outfile.Name,writeerr)
                }

                defer outfile.Close()
                cmd.Stdout = outfile

		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		if startErr := cmd.Start(); err != nil {
    			errorlog.Println("Could not start the enrolld script",startErr)
                        return startErr
		}

		timer := time.AfterFunc(time.Duration(configuration.Timeout) * time.Second, func() {
			errorlog.Println("The server " + server.FQDN + " have reached the timeout - Killing process",cmd.Process.Pid)
			pgid, err := syscall.Getpgid(cmd.Process.Pid)
			if err == nil {
    				syscall.Kill(-pgid, 15)
			}
		})
	
		execErr := cmd.Wait()
		timer.Stop()

		if execErr != nil {
			errorlog.Println("Error while excecuting script. Please see the log for more info: " + configuration.LogPath + "/" + server.FQDN + ".log")
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
                                	httpGetInventory(w,r)
				case "/targets":
                                	httpGetTargets(w,r)
				default:
                                	fmt.Fprintf(w, "running")
			}
                case "POST":
                        httpPost(w,r)
        }
}


func httpGetInventory(w http.ResponseWriter, r *http.Request) {
       	inventories, inventorieserr := GetInventory(configuration.Path)

        if inventorieserr != nil {
                errorlog.Println(inventorieserr)
                http.Error(w,http.StatusText(500),500)
        }

        inventory, inventoryErr := GetInventoryInJSON(inventories)
        if inventoryErr != nil {
                errorlog.Println("Error")
                http.Error(w,http.StatusText(500),500)
                return
        }

        fmt.Fprintf(w,inventory)
}


func httpGetTargets(w http.ResponseWriter, r *http.Request) {
       	servers, getErr := GetInventory(configuration.Path)

        if getErr != nil {
                errorlog.Println(getErr)
                http.Error(w,http.StatusText(500),500)
        }

        targets, targetsErr := GetTargetsInJSON(servers)
        if targetsErr != nil {
                errorlog.Println(targetsErr)
                http.Error(w,http.StatusText(500),500)
                return
        }

        fmt.Fprintf(w,targets)
}


func httpPost(w http.ResponseWriter, r *http.Request) {
        requestIP, _, err := net.SplitHostPort(r.RemoteAddr)
        if err == nil {
		var server ServerInfo

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
			server.Inventories = strings.Split(r.FormValue("inventory"),",")
			
			if strings.TrimSpace(r.FormValue("AnsibleProperties")) != "" {
				if m,_ := regexp.MatchString("^([: ,_'\"a-zA-Z0-9]*)$", r.FormValue("AnsibleProperties")); !m {
					errorlog.Println("Received AnsibleProperties with illegal characters: \"" + r.FormValue("AnsibleProperties") + "\" (",requestIP,")")
				} else {
  					jsonproperties := "{" + r.FormValue("AnsibleProperties") + " }"
                                        jsonproperties = strings.Replace(jsonproperties,"'","\"",-1)

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

		if (server.FQDN == "") {
			addresses, err := net.LookupAddr(requestIP)
    			if err == nil && len(addresses) >= 1 {
				addr := addresses[0]
				if addr != "" {
					addr = strings.TrimSuffix(addr,".")
					infolog.Println("FQDN was empty (" + requestIP + ") but the IP had the following name: \"" + addr + "\"")
					server.FQDN = addr
				}
			} else {
      				errorlog.Println("FQDN is empty (",requestIP,")")
                        	http.Error(w,http.StatusText(500),500)
				return
			}
		}

   		if m,_ := regexp.MatchString("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$", server.FQDN); !m {
                        errorlog.Println("Received FQDN with illegal characters: \"", server.FQDN,"\" (",requestIP,")")
                       	http.Error(w,http.StatusText(500),500)
                        return
                }

		if len(strings.Split(server.FQDN,".")) < 3 {
			addresses, err := net.LookupAddr(requestIP)
                        if err == nil && len(addresses) >= 1 {
				addr := addresses[0]
				if addr != "" {
					addr = strings.TrimSuffix(addr,".")
					infolog.Println("Server \"" + server.FQDN + "\"'s domain looks wrong - Replacing it with \"" + addr + "\"")
					server.FQDN = addr
				}
                        } else {
                        	errorlog.Println("Server \"" + server.FQDN + "\"'s domain looks wrong, but no suitable name was found to replace it")
                      	}
		}


                server.IP = requestIP
                server.LastSeen = time.Now().String()

		for _, fqdn := range configuration.Blacklist {
                        if strings.ToLower(server.FQDN) == strings.ToLower(fqdn) {
			   	infolog.Println(server.FQDN + " (" + server.IP + ") is on the blacklist - Ignoring")
	                        fmt.Fprintln(w, "Ignored")
        	                return
                        }
                }

		if strings.ToLower(r.FormValue("NewServer")) == "true" || strings.ToLower(r.FormValue("NewServer")) == "yes"  || strings.ToLower(server.NewServer) == "true" || strings.ToLower(server.NewServer) == "yes" {
			isNewServer = true
		}

		if (!server.Exist() || isNewServer) {
			isNewServer = true
			
                        enrolldErr := callEnrolldScript(server)
                        if enrolldErr != nil {
                       		errorlog.Println("Error running script against " + server.FQDN + "(" + server.IP + ")" + ": " + enrolldErr.Error())

				if val, ok := server.AnsibleProperties["global_server_type"]; !ok {
					notification("Enrolld failure", "Failed to enroll the following new server: " + server.FQDN + "(" + server.IP + ")", server)
				} else {
					if val != "clud" {
						notification("Enrolld failure", "Failed to enroll the following new server: " + server.FQDN + "(" + server.IP + ")",server)
					}
				}

                        	http.Error(w,http.StatusText(500),500)
				return
                        } else {
                        	infolog.Println("Enrolld script successful: " + server.FQDN)
                 	}
		}

                var writeerr error
                writeerr = WriteToFile(server, configuration.Path + "/" + server.FQDN, false)

                if (writeerr != nil){
                        errorlog.Println("Write Error")
                        http.Error(w,http.StatusText(500),500)
                } else {
                        fmt.Fprintln(w, "Enrolled")

                        if isNewServer {
             			infolog.Println("Enrolled the following new machine: " + server.FQDN + " (" + server.IP + ")")
			} else {
                                infolog.Println("Updated the following machine: " + server.FQDN + " (" + server.IP + ")")
			}
                }
        }
}


func initializeLogging(infologHandle io.Writer, errorlogHandle io.Writer) {
	infolog = log.New(infologHandle, "INFO: ", log.Lshortfile)
	errorlog = log.New(errorlogHandle, "ERROR: ", log.Lshortfile)
	infolog.Println("Logging Initialized")
}


func notification(subject string, message string, server ServerInfo) {
	binary, err := exec.LookPath(configuration.NotificationScriptPath)
	if err != nil {
		errorlog.Println("Could not find the notification script in the given path", configuration.NotificationScriptPath, err)
	}
	cmd := exec.Command(binary)

	env := os.Environ()
	env = append(env, fmt.Sprintf("SUBJECT=%s",subject))
	env = append(env, fmt.Sprintf("MESSAGE=%s",message))

	env = append(env, fmt.Sprintf("SERVER_FQDN=%s",server.FQDN))
	env = append(env, fmt.Sprintf("SERVER_IP=%s",server.IP))
	env = append(env, fmt.Sprintf("SERVER_PROPERTIES=%s",server.AnsibleProperties))
	env = append(env, fmt.Sprintf("SERVER_INVENTORIES=%s",server.Inventories))
	env = append(env, fmt.Sprintf("SERVER_LASTSEEN=%s",server.LastSeen))

	cmd.Env = env

	startErr := cmd.Start()
	if startErr != nil {
		errorlog.Println("Could not send notification", startErr)
	}

	cmd.Wait()
}


func initializeConfiguration(path string) {
        file,_ := os.Open(path)
        decoder := json.NewDecoder(file)
        err := decoder.Decode(&configuration)

        if err != nil {
                errorlog.Println("Error while reading the configuration file - Exiting")
                errorlog.Println(err)
                os.Exit(1)
        }

        _, existsErr := os.Stat(configuration.Path)

        if os.IsNotExist(existsErr) {
                createErr := os.MkdirAll(configuration.Path,0744)
                if createErr != nil {
                        errorlog.Println(createErr)
                } else {
                        infolog.Println("Created: " + configuration.Path)
                }
        }
}


func main() {
	initializeLogging(os.Stdout, os.Stderr)
	initializeConfiguration("/etc/enrolld/enrolld.conf")

      	scriptPathErr := checkScriptPath()
        if scriptPathErr != nil {
                log.Fatal("ScriptPath Problem - stopping")
        }

	infolog.Println("Listening on port: " + configuration.Port + " (http) and port: " + configuration.TlsPort + " (https)")
        http.HandleFunc("/", httpHandler)

	go http.ListenAndServe(":" + configuration.Port, nil)
	tlserr := http.ListenAndServeTLS(":" + configuration.TlsPort, configuration.TlsCert, configuration.TlsKey, nil)
	if tlserr != nil {
		errorlog.Println("Error starting TLS: ",tlserr)
	}
	errorlog.Println("Error happend while serving port")	
}
