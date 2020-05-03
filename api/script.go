package api

import (
	"fmt"
	"net"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/auth"
	"github.com/Olling/Enrolld/input"
	"github.com/Olling/Enrolld/config"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/output"
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

	script,ok := utils.Scripts[scriptID]
	if !ok {
		slog.PrintError("The script", scriptID, "requested by", requestIP, "was not found in memory")
		http.Error(w, "Please provide a valid ScriptID", 400)
		return
	}

	server, err := output.GetServer(serverID)
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

	err = input.RunScript(script.Path, server, scriptID, script.Timeout)
	if err != nil {
		slog.PrintError("The script", scriptID, "requested by", requestIP, "returned an error", err)
		http.Error(w, "Failed to run script", 500)
		return
	}

	slog.PrintInfo("The script", scriptID, "requested by", requestIP, "was successfully run against", serverID)
}


func getScripts(w http.ResponseWriter, r *http.Request) {
	if utils.Scripts == nil {
		fmt.Fprintln(w, "{}")
		return
	}

	json, err := utils.StructToJson(utils.Scripts)

	if err != nil {
		slog.PrintError("Failed to convert Scripts to json", err)
	}

	fmt.Fprintln(w, json)
}
