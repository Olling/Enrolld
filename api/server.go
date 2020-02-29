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
	input "github.com/Olling/Enrolld/input"
)


func updateServer(w http.ResponseWriter, r *http.Request) {
	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	contentType := r.Header.Get("Content-type")
	var servername string

	if r.FormValue("FQDN") != "" {
		servername = r.FormValue("FQDN")
	}

	if r.Header.Get("FQDN") != "" {
		servername = r.Header.Get("FQDN")
	}

	var server utils.ServerInfo

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

	servername, err = input.VerifyFQDN(server.ServerID, requestIP)
	server.ServerID = servername

	for _, fqdn := range config.Configuration.Blacklist {
		if strings.ToLower(server.ServerID) == strings.ToLower(fqdn) {
			slog.PrintDebug(server.ServerID + " (" + server.IP + ") is on the blacklist - Ignoring")
			fmt.Fprintln(w, "Ignored")
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
	var servername string

	params := mux.Vars(r)
	servername = params["servername"]

	if servername == "" {
		keys, ok := r.URL.Query()["servername"]

		if ok && len(keys) == 1 {
			servername = r.URL.Query()["servername"][0]
		}
	}

	server, err := output.GetServer(servername)

	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	serverjson, err := utils.StructToJson(server)
	fmt.Fprintln(w, serverjson)
}


func deleteServer(w http.ResponseWriter, r *http.Request) {
	var servername string

	params := mux.Vars(r)
	servername = params["servername"]

	if servername == "" {
		keys, ok := r.URL.Query()["servername"]

		if ok && len(keys) == 1 {
			servername = r.URL.Query()["servername"][0]
		}
	}

	_, err := output.GetServer(servername)

	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	err = fileio.DeleteServer(servername)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	metrics.ServersDeleted.Inc()
}
