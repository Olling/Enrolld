package main

import (
	//"encoding/json"
	//"fmt"
	//"io/ioutil"
	//"log"
	//"net"
	//"net/http"
	//"os/exec"
	//"os"
	//"regexp"
	//"strings"
	//"syscall"
	//"time"
	"os"
	"log"
	"github.com/Olling/Enrolld/api"
	"github.com/Olling/Enrolld/fileio"
	"github.com/Olling/Enrolld/config"
	l "github.com/Olling/Enrolld/logging"
)



//func httpPost(w http.ResponseWriter, r *http.Request) {
//	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)
//	if err == nil {
//		var server utils.ServerInfo
//
//		if r.FormValue("FQDN") != "" {
//			server.FQDN = r.FormValue("FQDN")
//		}
//
//		if r.Header.Get("FQDN") != "" {
//			server.FQDN = r.Header.Get("FQDN")
//		}
//
//		contentType := r.Header.Get("Content-type")
//		switch contentType {
//		case "application/json":
//			if r.Body == nil {
//				http.Error(w, "Please send a request body in JSON format", 400)
//				return
//			}
//
//			err := json.NewDecoder(r.Body).Decode(&server)
//			if err != nil {
//				http.Error(w, "The received JSON body was in the wrong format", 400)
//				return
//			}
//		default:
//			server.Inventories = strings.Split(r.FormValue("inventory"), ",")
//
//			if strings.TrimSpace(r.FormValue("AnsibleProperties")) != "" {
//				if m, _ := regexp.MatchString("^([: ,_'\"a-zA-Z0-9]*)$", r.FormValue("AnsibleProperties")); !m {
//					l.ErrorLog.Println("Received AnsibleProperties with illegal characters: \""+r.FormValue("AnsibleProperties")+"\" (", requestIP, ")")
//				} else {
//					jsonproperties := "{" + r.FormValue("AnsibleProperties") + " }"
//					jsonproperties = strings.Replace(jsonproperties, "'", "\"", -1)
//
//					var ansibleProperties map[string]string
//
//					err := json.NewDecoder(strings.NewReader(jsonproperties)).Decode(&ansibleProperties)
//					if err != nil {
//						http.Error(w, "The AnsibleProperties was in the wrong format", 400)
//						return
//					}
//
//					server.AnsibleProperties = ansibleProperties
//				}
//			}
//		}
//
//		isNewServer := false
//
//		if server.FQDN == "" {
//			addresses, err := net.LookupAddr(requestIP)
//			if err == nil && len(addresses) >= 1 {
//				addr := addresses[0]
//				if addr != "" {
//					addr = strings.TrimSuffix(addr, ".")
//					l.InfoLog.Println("FQDN was empty (" + requestIP + ") but the IP had the following name: \"" + addr + "\"")
//					server.FQDN = addr
//				}
//			} else {
//				l.ErrorLog.Println("FQDN is empty (", requestIP, ")")
//				http.Error(w, http.StatusText(500), 500)
//				return
//			}
//		}
//
//		if m, _ := regexp.MatchString("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$", server.FQDN); !m {
//			l.ErrorLog.Println("Received FQDN with illegal characters: \"", server.FQDN, "\" (", requestIP, ")")
//			http.Error(w, http.StatusText(500), 500)
//			return
//		}
//
//		if len(strings.Split(server.FQDN, ".")) < 3 {
//			addresses, err := net.LookupAddr(requestIP)
//			if err == nil && len(addresses) >= 1 {
//				addr := addresses[0]
//				if addr != "" {
//					addr = strings.TrimSuffix(addr, ".")
//					l.InfoLog.Println("Server \"" + server.FQDN + "\"'s domain looks wrong - Replacing it with \"" + addr + "\"")
//					server.FQDN = addr
//				}
//			} else {
//				l.ErrorLog.Println("Server \"" + server.FQDN + "\"'s domain looks wrong, but no suitable name was found to replace it")
//			}
//		}
//
//		server.IP = requestIP
//		server.LastSeen = time.Now().String()
//
//		for _, fqdn := range config.Configuration.Blacklist {
//			if strings.ToLower(server.FQDN) == strings.ToLower(fqdn) {
//				l.InfoLog.Println(server.FQDN + " (" + server.IP + ") is on the blacklist - Ignoring")
//				fmt.Fprintln(w, "Ignored")
//				return
//			}
//		}
//
//		if strings.ToLower(r.FormValue("NewServer")) == "true" || strings.ToLower(r.FormValue("NewServer")) == "yes" || strings.ToLower(server.NewServer) == "true" || strings.ToLower(server.NewServer) == "yes" {
//			isNewServer = true
//		}
//
//		if !server.Exist() || isNewServer {
//			isNewServer = true
//
//			enrolldErr := callEnrolldScript(server)
//			if enrolldErr != nil {
//				l.ErrorLog.Println("Error running script against " + server.FQDN + "(" + server.IP + ")" + ": " + enrolldErr.Error())
//
//				if val, ok := server.AnsibleProperties["global_server_type"]; !ok {
//					notification("Enrolld failure", "Failed to enroll the following new server: "+server.FQDN+"("+server.IP+")", server)
//				} else {
//					if val != "clud" {
//						notification("Enrolld failure", "Failed to enroll the following new server: "+server.FQDN+"("+server.IP+")", server)
//					}
//				}
//
//				http.Error(w, http.StatusText(500), 500)
//				return
//			} else {
//				l.InfoLog.Println("Enrolld script successful: " + server.FQDN)
//			}
//		}
//
//		var writeerr error
//		writeerr = io.WriteToFile(server, config.Configuration.Path+"/"+server.FQDN, false)
//
//		if writeerr != nil {
//			l.ErrorLog.Println("Write Error")
//			http.Error(w, http.StatusText(500), 500)
//		} else {
//			fmt.Fprintln(w, "Enrolled")
//
//			if isNewServer {
//				l.InfoLog.Println("Enrolled the following new machine: " + server.FQDN + " (" + server.IP + ")")
//			} else {
//				l.InfoLog.Println("Updated the following machine: " + server.FQDN + " (" + server.IP + ")")
//			}
//		}
//	}
//}

func main() {
	l.InitializeLogging(os.Stdout, os.Stderr)
	config.InitializeConfiguration("/etc/enrolld/enrolld.conf")

	scriptPathErr := fileio.CheckScriptPath()
	if scriptPathErr != nil {
		log.Fatal("ScriptPath Problem - stopping")
	}

	//metrics.Init()
	api.SetupRouter()
}
