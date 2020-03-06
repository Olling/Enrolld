package api

import (
	"fmt"
	"net/http"
)

func getOverwrites(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "GET LABEL!")
}

func getOverwrite(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "GET LABEL!")
}

func addOverwrite(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ADD LABEL!")
}

func deleteOverwrite(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "GET LABEL!")
}

