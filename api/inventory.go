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

func getInventory(w http.ResponseWriter, r *http.Request) {
	requestIP, _, err := net.SplitHostPort(r.RemoteAddr)
	servers, err := dataaccess.GetServers()

	if !auth.CheckAccess(w,r, "read", objects.Server{}) {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintError("Unauthorized call to get inventory from", requestIP)
		return
	}

	if err != nil {
		slog.PrintError(err)
		http.Error(w, http.StatusText(500), 500)
	}

	inventory, err := output.GetInventoryInJSON(servers)
	if err != nil {
		slog.PrintError(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Fprintf(w, inventory)
}
