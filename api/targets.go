package api

import (
	"fmt"
	"net/http"
	"github.com/Olling/Enrolld/output"
)

func getTargets(w http.ResponseWriter, r *http.Request) {
	servers, getErr := output.GetInventory()

	if getErr != nil {
		fmt.Println(getErr)
		http.Error(w, http.StatusText(500), 500)
	}

	targets, targetsErr := output.GetTargetsInJSON(servers)
	if targetsErr != nil {
		fmt.Println(targetsErr)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Fprintf(w, targets)
}
