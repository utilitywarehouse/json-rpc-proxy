package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/utilitywarehouse/uw-bill-rpc-handler/extpoints"
)

const n = "default"

func init() {
	log.Print("registering linechecker handler")
	ok := endpoints.Register(extpoints.Endpoint(func() http.HandlerFunc { return t }), n)
	if !ok {
		log.Panicf("handler name: %s failed to register", n)
	}
}

func t(wr http.ResponseWriter, req *http.Request) {
	fmt.Fprint(wr, "hello")
}
