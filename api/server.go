package api

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"net/http"
	"io/ioutil"
	"github.com/gorilla/mux"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/auth"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/dataaccess"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enrolld/dataaccess/config"
)


func updateServer(w http.ResponseWriter, r *http.Request) {
	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		slog.PrintError("Failed to get request IP")
		return
	}

	contentType := r.Header.Get("Content-type")
	var serverID string

	if r.FormValue("FQDN") != "" {
		serverID = r.FormValue("FQDN")
	}

	if r.Header.Get("FQDN") != "" {
		serverID = r.Header.Get("FQDN")
	}

	var server objects.Server

	if contentType == "application/json" {
		if r.Body == nil {
			slog.PrintError("Empty POST received from:" , requestIP)
			http.Error(w, "Please send a request body in JSON format", 400)
			return
		}

		json, err := ioutil.ReadAll(r.Body)
		if err != nil {
			slog.PrintError("Could not read body from:" , requestIP)
			http.Error(w, "Please send a request body in JSON format", 400)
			return
		}

		err = utils.StructFromJson(json, &server)
		if err != nil {
			slog.PrintError("Could not decode JSON from:" , requestIP, err)
			http.Error(w, "The received JSON body was in the wrong format", 400)
			return
		}
	}

	serverID, err = utils.VerifyFQDN(server.ServerID, requestIP)
	server.ServerID = serverID
	server.IP = requestIP

	for _,regexString := range config.Configuration.Blacklist {
		slog.PrintDebug("Found regular expression:", regexString)
		if match,_ := regexp.MatchString(regexString, server.ServerID); match {
			slog.PrintInfo(server.ServerID + " (" + server.IP + ") is on the blacklist - Ignoring")
			slog.PrintDebug(server.ServerID + " (" + server.IP + ") matched the following blacklist regular expression:", regexString)
			fmt.Fprintln(w, "Server is blacklisted - Ignoring")
			return
		}
	}

	if !auth.CheckAccess(w,r, "write", server) {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintError("Unauthorized call to update", server.ServerID, "from", requestIP)
		return
	}

	isNewServer := false

	//Enroll server if it does not exist
	if !dataaccess.ServerExist(server) {
		isNewServer = true
	}

	//Enroll server if overwrite is found
	if strings.ToLower(r.FormValue("NewServer")) == "true" || strings.ToLower(r.FormValue("NewServer")) == "yes" || server.NewServer {
		isNewServer = true
	}

	slog.PrintDebug(isNewServer)
	if isNewServer {
		dataaccess.EnrollServer(server)
	} else {
		dataaccess.UpdateServer(server)
	}
}


func getServer(w http.ResponseWriter, r *http.Request) {
	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)
	var serverID string

	params := mux.Vars(r)
	serverID = params["serverid"]

	if serverID == "" {
		keys, ok := r.URL.Query()["serverid"]

		if ok && len(keys) == 1 {
			serverID = r.URL.Query()["serverid"][0]
		}
	}

	server, err := dataaccess.GetServer(serverID)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	if !auth.CheckAccess(w, r, "read", server) {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintError("Unauthorized call to read", server.ServerID, "from", requestIP)
		return
	}

	serverjson, err := utils.StructToJson(server)
	fmt.Fprintln(w, serverjson)
}


func deleteServer(w http.ResponseWriter, r *http.Request) {
	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)
	var serverID string

	params := mux.Vars(r)
	serverID = params["serverid"]

	if serverID == "" {
		keys, ok := r.URL.Query()["serverid"]

		if ok && len(keys) == 1 {
			serverID = r.URL.Query()["serverid"][0]
		}
	}

	server, err := dataaccess.GetServer(serverID)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	if !auth.CheckAccess(w, r, "write", server) {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintError("Unauthorized call to delete", server.ServerID, "from", requestIP)
		return
	}

	err = dataaccess.RemoveServer(serverID)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
}
