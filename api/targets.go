package api

import (
	"fmt"
	"net/http"
)

func getTargets(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "GET Targets!")
}