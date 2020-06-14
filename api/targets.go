package api

import (
	"fmt"
	"net"
	"net/http"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/auth"
	"github.com/Olling/Enrolld/output"
	"github.com/Olling/Enrolld/dataaccess"
	"github.com/Olling/Enrolld/utils/objects"
)

func getTargets(w http.ResponseWriter, r *http.Request) {
	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if !auth.CheckAccess(w,r, "read", objects.Server{}) {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintError("Unauthorized call to get targets from", requestIP)
		return
	}

	servers, err := dataaccess.GetServers()
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
	}

	targets, err := output.GetTargetsInJSON(servers)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, targets)
}
