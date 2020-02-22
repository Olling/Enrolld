package api

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// SetupRouter initializes the API routes
func SetupRouter() {
	router := mux.NewRouter()
	router.HandleFunc("/", rootHandler)
	router.HandleFunc("/server", getServer).Methods("GET")
	router.HandleFunc("/server", addServer).Methods("POST")
	router.HandleFunc("/server", updateServer).Methods("PUT")
	router.HandleFunc("/server", deleteServer).Methods("DELETE")

	loggedRouter := handlers.CombinedLoggingHandler(os.Stdout, router)
	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
}

func rootHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}
