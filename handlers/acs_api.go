package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/utilitywarehouse/cpe-configurations-api/client"
	acs "github.com/utilitywarehouse/cpe-configurations-api/model"
	"github.com/utilitywarehouse/json-rpc-proxy/extpoints"
)

const (
	acsApiSearchRoute = "acs/search"
	acsApiUpdateRoute = "acs/update"
	acsApiCreateRoute = "acs/create"
)

var (
	acsApiHostname = "cpe-configurations-api.%s.uw.systems"
	acsClient      = cpeConfClient.NewInstrumentedApiClient(cpeConfClient.NewHttpApiClient(fmt.Sprintf(acsApiHostname, env)))
)

func init() {
	log.Print("registering ACS search route")
	ok := endpoints.Register(extpoints.Endpoint(func() http.HandlerFunc { return handleSearchConfigurations }), acsApiSearchRoute)
	if !ok {
		log.Panicf("handler name: %s failed to register", route)
	}
	ok = endpoints.Register(extpoints.Endpoint(func() http.HandlerFunc { return handleUpdateConfiguration }), acsApiUpdateRoute)
	if !ok {
		log.Panicf("handler name: %s failed to register", route)
	}
	ok = endpoints.Register(extpoints.Endpoint(func() http.HandlerFunc { return handleCreateConfiguration }), acsApiCreateRoute)
	if !ok {
		log.Panicf("handler name: %s failed to register", route)
	}
}

func handleUpdateConfiguration(wr http.ResponseWriter, req *http.Request) {
	conf, err := decodeUpdateConfiguration(req.Body)
	if err != nil {
		http.Error(wr, fmt.Sprintf("Could not decode incoming request: %v", err), http.StatusBadRequest)
		return
	}
	err = acsClient.UpdateConfiguration(conf)
	if err != nil {
		http.Error(wr, fmt.Sprintf("Error received from upstream service: %v", err), http.StatusBadGateway)
		return
	}
}

func decodeUpdateConfiguration(body io.ReadCloser) (acs.Configuration, error) {
	conf := acs.Configuration{}
	reqBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return conf, fmt.Errorf("Could not read request body: %v", err)
	}
	err = json.Unmarshal(reqBytes, &conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

func handleSearchConfigurations(wr http.ResponseWriter, req *http.Request) {
	accountNumber, cli := decodeSearchQuery(req)
	confQuery := acs.ConfigurationQuery{}
	if len(accountNumber) > 0 {
		confQuery.AccountNumber = accountNumber
	}
	if len(cli) > 0 {
		confQuery.Username = cli + "@uwclub.net"
	}

	res, err := acsClient.SearchConfigurations(confQuery)
	if err != nil {
		http.Error(wr, fmt.Sprintf("Error received from upstream service: %+v", err), http.StatusBadGateway)
		return
	}
	if len(res) > 1 {
		log.Printf("Received more than 1 configuration for query, using first of: %s", spew.Sdump(res))
	}
	encoder := json.NewEncoder(wr)
	if len(res) == 0 {
		if err = encoder.Encode(acs.Configuration{}); err != nil {
			log.Printf("Error encoding response to downstream service: %v", err)
		}
		return
	}
	if err = encoder.Encode(res[0]); err != nil {
		log.Printf("Error encoding response to downstream service: %v", err)
	}
}

func decodeSearchQuery(req *http.Request) (string, string) {
	queryString := req.URL.Query()
	cli := queryString.Get("cli")
	accountNumber := queryString.Get("accountNumber")
	return accountNumber, cli
}

func handleCreateConfiguration(wr http.ResponseWriter, req *http.Request) {
	reqBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(wr, fmt.Sprintf("Could not decode incoming request: %v", err), http.StatusBadRequest)
		return
	}
	conf := &acs.Configuration{}
	err = json.Unmarshal(reqBytes, conf)
	if err != nil {
		if err != nil {
			http.Error(wr, fmt.Sprintf("Could not decode incoming request: %v", err), http.StatusBadRequest)
			return
		}
	}
	err = acsClient.CreateConfiguration(*conf)
	if err != nil {
		if err != nil {
			http.Error(wr, fmt.Sprintf("Error received from upstream service: %v", err), http.StatusBadGateway)
			return
		}
	}
}
