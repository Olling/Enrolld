package api

import (
	"fmt"
	"net"
	"strings"
	"net/http"
	"io/ioutil"
	"github.com/gorilla/mux"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/fileio"
	"github.com/Olling/Enrolld/output"
	"github.com/Olling/Enrolld/config"
	"github.com/Olling/Enrolld/metrics"
	"github.com/Olling/Enrolld/input"
)


func updateServer(w http.ResponseWriter, r *http.Request) {
	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
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

	var server utils.Server

	switch contentType {
	case "application/json":
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

	serverID, err = input.VerifyFQDN(server.ServerID, requestIP)
	server.ServerID = serverID

	for _, fqdn := range config.Configuration.Blacklist {
		if strings.ToLower(server.ServerID) == strings.ToLower(fqdn) {
			slog.PrintDebug(server.ServerID + " (" + server.IP + ") is on the blacklist - Ignoring")
			fmt.Fprintln(w, "Server is blacklisted - Ignoring")
			return
		}
	}

	isNewServer := false
	if strings.ToLower(r.FormValue("NewServer")) == "true" || strings.ToLower(r.FormValue("NewServer")) == "yes" || server.NewServer {
		isNewServer = true
	}

	input.UpdateServer(server, isNewServer)

	metrics.ServersUpdated.Inc()
}


func getServer(w http.ResponseWriter, r *http.Request) {
	var serverID string

	params := mux.Vars(r)
	serverID = params["serverid"]

	if serverID == "" {
		keys, ok := r.URL.Query()["serverid"]

		if ok && len(keys) == 1 {
			serverID = r.URL.Query()["serverid"][0]
		}
	}

	server, err := output.GetServer(serverID)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	serverjson, err := utils.StructToJson(server)
	fmt.Fprintln(w, serverjson)
}


func deleteServer(w http.ResponseWriter, r *http.Request) {
	var serverID string

	params := mux.Vars(r)
	serverID = params["serverid"]

	if serverID == "" {
		keys, ok := r.URL.Query()["serverid"]

		if ok && len(keys) == 1 {
			serverID = r.URL.Query()["serverid"][0]
		}
	}

	_, err := output.GetServer(serverID)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	err = fileio.DeleteServer(serverID)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	metrics.ServersDeleted.Inc()
}
