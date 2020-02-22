package api

import (
	"fmt"
	"net/http"

	"github.com/Olling/Enrolld/output"
)

func getTargets(w http.ResponseWriter, r *http.Request) {
	servers, err := output.GetInventory()

	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), 500)
	}

	targets, targetsErr := output.GetTargetsInJSON(servers)
	if err = nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Fprintf(w, targets)
}
