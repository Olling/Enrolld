package api

import (
	"fmt"
	"net"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/auth"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/dataaccess"
	"github.com/Olling/Enrolld/dataaccess/config"
)


func runScript(w http.ResponseWriter, r *http.Request) {
	requestIP, _,_ := net.SplitHostPort(r.RemoteAddr)
	var serverID string
	var scriptID string

	params := mux.Vars(r)
	serverID = params["serverid"]
	scriptID = params["scriptid"]


	if serverID == "" {
		keys, ok := r.URL.Query()["serverid"]

		if ok && len(keys) == 1 {
			serverID = r.URL.Query()["serverid"][0]
		}
	}

	if scriptID == "" {
		keys, ok := r.URL.Query()["scriptid"]

		if ok && len(keys) == 1 {
			serverID = r.URL.Query()["serverid"][0]
		}
	}

	if scriptID == "" || !utils.ValidInput(scriptID) {
		slog.PrintError("Invalid or empty ScriptID was provided by:", requestIP)
		http.Error(w, "Please provide a valid ScriptID", 400)
		return
	}

	if serverID == "" || !utils.ValidInput(serverID) {
		slog.PrintError("Invalid or empty ServerID was provided by:", requestIP)
		http.Error(w, "Please provide a valid ServerID", 400)
		return
	}

	script,ok := dataaccess.Scripts[scriptID]
	if !ok {
		slog.PrintError("The script", scriptID, "requested by", requestIP, "was not found in memory")
		http.Error(w, "Please provide a valid ScriptID", 400)
		return
	}

	server, err := dataaccess.GetServer(serverID)
	if err != nil {
		slog.PrintError("The script", scriptID, "requested by", requestIP, "got a server ID that was not enrolled:", serverID)
		http.Error(w, "Please provide a valid ServerID", 400)
		return
	}

	if !auth.CheckAccess(w,r, "execute", server) {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintError("Unauthorized call to run script", scriptID ,"from", requestIP)
		return
	}

	if script.Timeout == 0 {
		script.Timeout = config.Configuration.Timeout
	}

	err = dataaccess.RunScript(script.Path, server, scriptID, script.Timeout)
	if err != nil {
		slog.PrintError("The script", scriptID, "requested by", requestIP, "returned an error", err)
		http.Error(w, "Failed to run script", 500)
		return
	}

	slog.PrintInfo("The script", scriptID, "requested by", requestIP, "was successfully run against", serverID)
}


func getScripts(w http.ResponseWriter, r *http.Request) {
	if dataaccess.Scripts == nil {
		fmt.Fprintln(w, "{}")
		return
	}

	json, err := utils.StructToJson(dataaccess.Scripts)

	if err != nil {
		slog.PrintError("Failed to convert Scripts to json", err)
	}

	fmt.Fprintln(w, json)
}
