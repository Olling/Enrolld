package api

import (
	"fmt"
	"net/http"
)

func addLabel(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ADD LABEL!")
}

func getLabel(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "GET LABEL!")
}
