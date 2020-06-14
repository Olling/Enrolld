package api

import (
	"fmt"
	"net"
	"net/http"
	"io/ioutil"
	"github.com/gorilla/mux"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/dataaccess"
	"github.com/Olling/Enrolld/utils/objects"
)

func getOverwrites(w http.ResponseWriter, r *http.Request) {
	json,err := utils.StructToJson(dataaccess.Overwrites)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	fmt.Fprintln(w, json)
}

func getOverwrite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	overwriteID := params["overwriteid"]

	if overwriteID == "" {
		keys, ok := r.URL.Query()["overwriteid"]

		if ok && len(keys) == 1 {
			overwriteID = r.URL.Query()["overwriteid"][0]
		}
	}

	if _,ok := dataaccess.Overwrites[overwriteID]; ok {
		json,err := utils.StructToJson(dataaccess.Overwrites[overwriteID])
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		fmt.Fprintln(w, json)
	}
}

func addOverwrite(w http.ResponseWriter, r *http.Request) {
	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)
	var overwrite objects.Overwrite

	params := mux.Vars(r)
	overwriteID := params["overwriteid"]

	if overwriteID == "" {
		keys, ok := r.URL.Query()["overwriteid"]

		if ok && len(keys) == 1 {
			overwriteID = r.URL.Query()["overwriteid"][0]
		}
	}

	if overwriteID == "" {
		slog.PrintError("Did not get Overwrite ID from:" , requestIP)
		http.Error(w, "Please provide a Overwrite ID", 400)
		return
	}

	contentType := r.Header.Get("Content-type")
	if contentType != "application/json" {
		http.Error(w, http.StatusText(415), 415)
		return
	}

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

	err = utils.StructFromJson(json, &overwrite)
	if err != nil {
		slog.PrintError("Could not decode JSON from:" , requestIP, err)
		http.Error(w, "The received JSON body was in the wrong format", 400)
		return
	}

	dataaccess.Overwrites[overwriteID] = overwrite
	dataaccess.SaveOverwrites()
}

func deleteOverwrite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	overwriteID := params["overwriteid"]

	if overwriteID == "" {
		keys, ok := r.URL.Query()["overwriteid"]

		if ok && len(keys) == 1 {
			overwriteID = r.URL.Query()["overwriteid"][0]
		}
	}

	delete(dataaccess.Overwrites, overwriteID)
}
