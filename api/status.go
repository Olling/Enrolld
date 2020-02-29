package api

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Olling/Enrolld/output"
)


func getStatus(w http.ResponseWriter, r *http.Request) {
	var serverid string

	params := mux.Vars(r)
	serverid = params["serverid"]

	if serverid == "" {
		keys, ok := r.URL.Query()["serverid"]

		if ok && len(keys) == 1 {
			serverid = r.URL.Query()["serverid"][0]
		}
	}

	_, err := output.GetServer(serverid)

	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	//TODO write 202 code (enrolling)

	w.WriteHeader(http.StatusOK)
}
