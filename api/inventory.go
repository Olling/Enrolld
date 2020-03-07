package api

import (
	"fmt"
	"net/http"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/output"
)

func getInventory(w http.ResponseWriter, r *http.Request) {
	servers, err := output.GetServers()

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
