package api

import (
	"fmt"
	"net"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Olling/Enrolld/output"
	input "github.com/Olling/Enrolld/input"
)


func getStatus(w http.ResponseWriter, r *http.Request) {
	var serverid string
	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)

	params := mux.Vars(r)
	serverid = params["serverid"]

	if serverid == "" {
		keys, ok := r.URL.Query()["serverid"]

		if ok && len(keys) == 1 {
			serverid = r.URL.Query()["serverid"][0]
		}
	}

	serverid, err = input.VerifyFQDN(serverid, requestIP)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		fmt.Println(err)
		return
	}

	_, err = output.GetServer(serverid)

	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		fmt.Println(err)
		return
	}

	//TODO write 202 code (enrolling)

	w.WriteHeader(http.StatusOK)
}
