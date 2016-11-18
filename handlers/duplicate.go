package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/utilitywarehouse/uw-bill-rpc-handler/extpoints"
)

const handlerName = "default"

var endpoints = extpoints.Endpoints

func init() {
	log.Print("registering linechecker handler")
	ok := endpoints.Register(extpoints.Endpoint(func() http.HandlerFunc { return handle }), handlerName)
	if !ok {
		log.Panicf("handler name: %s failed to register", handlerName)
	}
}

func handle(wr http.ResponseWriter, req *http.Request) {
	fmt.Fprint(wr, "hello")
}
