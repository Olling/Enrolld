package api

import (
	"fmt"
	"net"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Olling/Enrolld/output"
	"github.com/Olling/Enrolld/input"
)


func getStatus(w http.ResponseWriter, r *http.Request) {
	var serverID string
	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)

	params := mux.Vars(r)
	serverID = params["serverid"]

	if serverID == "" {
		keys, ok := r.URL.Query()["serverid"]

		if ok && len(keys) == 1 {
			serverID = r.URL.Query()["serverid"][0]
		}
	}

	serverID, err = input.VerifyFQDN(serverID, requestIP)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		fmt.Println(err)
		return
	}

	_, err = output.GetServer(serverID)

	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		fmt.Println(err)
		return
	}

	//TODO write 202 code (enrolling)

	w.WriteHeader(http.StatusOK)
}
