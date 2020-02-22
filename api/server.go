package api

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Olling/Enrolld/metrics"
	"github.com/Olling/Enrolld/output"
	"github.com/Olling/Enrolld/utils"
)

func addServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ADD SERVER!")

	// example counter
	metrics.ServerAddedCounter.Inc()
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

	server,err := output.GetServer(servername)

	if (err != nil) {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	serverjson,err := utils.StructToJson(server)
	fmt.Fprintln(w, serverjson)
}

func updateServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "UPDATE SERVER!")
}

func deleteServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "DELETE SERVER!")
}
