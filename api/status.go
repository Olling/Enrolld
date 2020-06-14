package api

import (
	"net"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/auth"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/dataaccess"
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

	serverID, err = utils.VerifyFQDN(serverID, requestIP)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	if dataaccess.IsServerActive(serverID) {
		http.Error(w, http.StatusText(208), 208)
		return
	}

	server, err := dataaccess.GetServer(serverID)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	if !auth.CheckAccess(w,r, "read", server) {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintError("Unauthorized call to get status on server", server.ServerID, "from", requestIP)
		return
	}


	w.WriteHeader(http.StatusOK)
}
