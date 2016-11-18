package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/utilitywarehouse/uw-bill-rpc-handler/extpoints"
)

type broadbandAvailabilityRequest struct {
	Address struct {
		BuildingName   string `json:"buildingName"`
		BuildingNumber string `json:"buildingNumber"`
		PostTown       string `json:"postTown"`
		Postcode       string `json:"postcode"`
		Street         string `json:"street"`
		SubBuilding    string `json:"subBuilding"`
	} `json:"address"`
	Cli              string `json:"cli"`
	ProductRequested string `json:"productRequested"`
	Source           string `json:"source"`
}

const route = "getbroadbandavailability"

var broadbandAvailabilityUrl = "http://test.example.com"

var endpoints = extpoints.Endpoints

func init() {
	log.Print("registering linechecker handler")
	ok := endpoints.Register(extpoints.Endpoint(func() http.HandlerFunc { return handleGetBroadbandAvailability }), route)
	if !ok {
		log.Panicf("handler name: %s failed to register", route)
	}
}

func handleGetBroadbandAvailability(wr http.ResponseWriter, req *http.Request) {
	var b []byte
	buf := bytes.NewBuffer(b)
	_, err := io.Copy(buf, req.Body)
	if err != nil {
		http.Error(wr, fmt.Sprintf("error reading request body %+v", err), http.StatusBadRequest)
		return
	}

	var raw interface{}
	err = json.Unmarshal(buf.Bytes(), &raw)
	if err != nil {
		http.Error(wr, fmt.Sprintf("error unmarshalling json: %+v", err), http.StatusBadRequest)
		return
	}

	m := raw.(map[string]interface{})
	var request = new(broadbandAvailabilityRequest)
	addProperties(request, m)
	d, err := json.Marshal(request)
	if err != nil {
		http.Error(wr, "failed to serialize response", http.StatusInternalServerError)
	}
	log.Printf("request: %+v", request)
	cl := http.Client{}
	sr, err := cl.Post(broadbandAvailabilityUrl, "application/json", bytes.NewReader(d))
	if err != nil {
		http.Error(wr, fmt.Sprintf("error getting response from upstream service %+v", err), http.StatusBadGateway)
	}

	responseHeader := wr.Header()
	for k, v := range sr.Header {
		responseHeader.Set(k, strings.Join(v, ","))
	}

	_, err = io.Copy(wr, sr.Body)
	if err != nil {
		http.Error(wr, fmt.Sprintf("error writing to response body: %+v", err), http.StatusInternalServerError)
	}

}

func addProperties(r *broadbandAvailabilityRequest, props map[string]interface{}) {
	for k, v := range props {
		log.Printf("key: %s, type: %+v", k, v)
		switch p := v.(type) {
		case string:
			switch k {
			case "cli":
				r.Cli = p
			case "source":
				r.Source = p
			case "productRequested":
				r.ProductRequested = p
			case "postcode":
				r.Address.Postcode = p
			case "buildingName":
				r.Address.BuildingName = p
			case "buildingNumber":
				r.Address.BuildingNumber = p
			case "street":
				r.Address.Street = p
			case "postTown":
				r.Address.Street = p
			}
		}
	}
}
