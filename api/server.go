package api

import (
	"fmt"
	"net/http"

	"github.com/Olling/Enrolld/metrics"
)

func addServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ADD SERVER!")

	// example counter
	metrics.ServerAddedCounter.Inc()
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
