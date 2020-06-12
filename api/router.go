package api

import (
	"os"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Olling/slog"
	"github.com/gorilla/handlers"
	"github.com/Olling/Enrolld/dataaccess/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRouter initializes the API routes
func SetupRouter() {
	router := mux.NewRouter()

	router.HandleFunc("/", rootHandler)

	router.HandleFunc("/server", getServer).Methods("GET")
	router.HandleFunc("/server/{serverid}", getServer).Methods("GET")

	router.HandleFunc("/server", updateServer).Methods("PUT","POST")
	router.HandleFunc("/server/{serverid}", updateServer).Methods("PUT", "POST")

	router.HandleFunc("/server", deleteServer).Methods("DELETE")
	router.HandleFunc("/server/{serverid}", deleteServer).Methods("DELETE")

	router.HandleFunc("/status", getStatus).Methods("GET")
	router.HandleFunc("/status/{serverid}", getStatus).Methods("GET")


	router.HandleFunc("/overwrites", getOverwrites).Methods("GET")

	router.HandleFunc("/overwrite", getOverwrite).Methods("GET")
	router.HandleFunc("/overwrite/{overwriteid}", getOverwrite).Methods("GET")

	router.HandleFunc("/overwrite", addOverwrite).Methods("POST")
	router.HandleFunc("/overwrite/{overwriteid}", addOverwrite).Methods("POST")

	router.HandleFunc("/overwrite", deleteOverwrite).Methods("Delete")
	router.HandleFunc("/overwrite/{overwriteid}", deleteOverwrite).Methods("Delete")

	router.HandleFunc("/script", runScript).Methods("POST")
	router.HandleFunc("/script/{scriptid}/{serverid}", runScript).Methods("POST")
	router.HandleFunc("/scripts", getScripts).Methods("GET")

	router.HandleFunc("/targets", getTargets).Methods("GET")
	router.HandleFunc("/inventory", getInventory).Methods("GET")

	// add prometheus handler
	router.Handle("/metrics", promhttp.Handler())

	// enable logging
	loggedRouter := handlers.CombinedLoggingHandler(os.Stdout, router)

	slog.PrintInfo("Listening on port: " + config.Configuration.Port + " (http) and port: " + config.Configuration.TlsPort + " (https)")

	go http.ListenAndServe(":"+config.Configuration.Port, loggedRouter)
	err := http.ListenAndServeTLS(":"+config.Configuration.TlsPort, config.Configuration.TlsCert, config.Configuration.TlsKey, loggedRouter)
	if err != nil {
		slog.PrintError("Error starting TLS: ", err)
	}
}


func rootHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "Welcome to Enrolld")
}
