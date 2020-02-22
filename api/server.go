package api

import (
	"fmt"
	"net/http"
)

func addServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ADD SERVER!")
}

func getServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "GET SERVER!")
}

func updateServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "UPDATE SERVER!")
}

func deleteServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "DELETE SERVER!")
}
