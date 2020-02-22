package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Olling/Enrolld/config"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// SetupRouter initializes the API routes
func SetupRouter() {
	router := mux.NewRouter()

	// api routes
	router.HandleFunc("/", rootHandler)
	router.HandleFunc("/server", getServer).Methods("GET")
	router.HandleFunc("/server", addServer).Methods("POST")
	router.HandleFunc("/server", updateServer).Methods("PUT")
	router.HandleFunc("/server", deleteServer).Methods("DELETE")
	router.HandleFunc("/label", getLabel).Methods("GET")
	router.HandleFunc("/label", addLabel).Methods("POST")
	router.HandleFunc("/targets", getTargets).Methods("GET")
	router.HandleFunc("/inventory", rootHandler).Methods("GET")

	// enable logging
	loggedRouter := handlers.CombinedLoggingHandler(os.Stdout, router)

	// needs infolog
	fmt.Println("Listening on port: " + config.Configuration.Port + " (http) and port: " + config.Configuration.TlsPort + " (https)")

	go http.ListenAndServe(":"+config.Configuration.Port, loggedRouter)
	err := http.ListenAndServeTLS(":"+config.Configuration.TlsPort, config.Configuration.TlsCert, config.Configuration.TlsKey, loggedRouter)
	if err != nil {
		// needs proper handling
		fmt.Println("Error starting TLS: ", err)
	}
}

func rootHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}
