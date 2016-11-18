//go:generate go-extpoints
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/utilitywarehouse/uw-bill-rpc-handler/extpoints"
	_ "github.com/utilitywarehouse/uw-bill-rpc-handler/handlers"
)

var endpoints = extpoints.Endpoints

func main() {
	router := mux.NewRouter()
	endpointProviders := endpoints.All()
	log.Printf("handlers: %+v", endpointProviders)
	for route, endpointProvider := range endpointProviders {
		log.Printf("registered handler for route: %s", route)
		router.HandleFunc("/"+route, endpointProvider())
	}

	log.Printf("router: %+v", router)
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)

	log.Fatal(http.ListenAndServe(":8000", loggedRouter))
}
