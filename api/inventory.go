package api

import (
	"fmt"
	"net/http"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/output"
)

func getInventory(w http.ResponseWriter, r *http.Request) {
	inventories, inventorieserr := output.GetInventory()

	if inventorieserr != nil {
		slog.PrintError(inventorieserr)
		http.Error(w, http.StatusText(500), 500)
	}

	inventory, inventoryErr := output.GetInventoryInJSON(inventories)
	if inventoryErr != nil {
		slog.PrintError(inventoryErr)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Fprintf(w, inventory)
}
