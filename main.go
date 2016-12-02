//go:generate go-extpoints
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/utilitywarehouse/go-operational/op"
	"github.com/utilitywarehouse/uw-bill-rpc-handler/extpoints"
	_ "github.com/utilitywarehouse/uw-bill-rpc-handler/handlers"
	"time"
)

var endpoints = extpoints.Endpoints

func main() {

	oph := op.NewHandler(op.NewStatus(
		"bill-rpc-handler",
		"Handles outbound flat JSON HTTP calls from Bill, and passes the real work on to a subsequent service").
		AddOwner("telecoms team", "#telecom").
		AddLink("github", "https://github.com/utilitywarehouse/uw-bill-rpc-handler").
		ReadyAlways(),
	)

	router := mux.NewRouter()
	router.NewRoute().PathPrefix("/__/").Handler(oph)
	endpointProviders := endpoints.All()

	log.Printf("handlers: %+v", endpointProviders)

	histoOpts := prometheus.HistogramOpts{
		Name: "bill_rpc_handler_downstream_latencies",
		Help: "A labeled histogram of the downstream service latencies",
	}

	latenciesHisto := prometheus.NewHistogramVec(histoOpts, []string{"route"})
	prometheus.DefaultRegisterer.MustRegister(latenciesHisto)

	for route, ep := range endpointProviders {
		log.Printf("registered handler for route: %s", route)
		router.HandleFunc("/"+route, createInstrumentedHandler(route, ep(), latenciesHisto))
	}

	router.Handle("/metrics", promhttp.Handler())

	log.Printf("router: %+v", router)
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Fatal(http.ListenAndServe(":8000", loggedRouter))
}

func createInstrumentedHandler(route string, provider http.HandlerFunc, histo *prometheus.HistogramVec) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		now := time.Now()
		provider.ServeHTTP(w, req)
		seconds := float64(time.Since(now).Seconds())
		histo.WithLabelValues(route).Observe(seconds)
	}
}
