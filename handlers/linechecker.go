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

type incomingBroabdandAvailabilityRequest struct {
	BuildingName     string
	BuildingNumber   string
	PostTown         string
	Postcode         string
	Street           string
	SubBuilding      string
	Cli              string
	ProductRequested string
	Source           string
}

type broadbandAvailabilityRequest struct {
	Address          broabdandAvailabilityRequestAddress `json:"address"`
	Cli              string                              `json:"cli"`
	ProductRequested string                              `json:"productRequested"`
	Source           string                              `json:"source"`
}

type broabdandAvailabilityRequestAddress struct {
	BuildingName   string `json:"buildingName"`
	BuildingNumber string `json:"buildingNumber"`
	PostTown       string `json:"postTown"`
	Postcode       string `json:"postcode"`
	Street         string `json:"street"`
	SubBuilding    string `json:"subBuilding"`
}

const route = "getbroadbandavailability"

var broadbandAvailabilityURL = "http://test.example.com"

var endpoints = extpoints.Endpoints

func init() {
	log.Print("registering linechecker handler")
	ok := endpoints.Register(extpoints.Endpoint(func() http.HandlerFunc { return handleGetBroadbandAvailability }), route)
	if !ok {
		log.Panicf("handler name: %s failed to register", route)
	}
}

func handleGetBroadbandAvailability(wr http.ResponseWriter, req *http.Request) {

	buf, err := getServiceRequest(req.Body)
	if err != nil {
		http.Error(wr, fmt.Sprintf("error converting request to upstream %+v", err), http.StatusBadRequest)
	}

	cl := http.Client{}
	sr, err := cl.Post(broadbandAvailabilityURL, "application/json", buf)
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

func getServiceRequest(body io.ReadCloser) (io.Reader, error) {
	dec := json.NewDecoder(body)
	var i incomingBroabdandAvailabilityRequest
	err := dec.Decode(i)
	if err != nil {
		return nil, err
	}

	out := &broadbandAvailabilityRequest{
		Cli:              i.Cli,
		ProductRequested: i.ProductRequested,
		Source:           i.Source,
		Address: broabdandAvailabilityRequestAddress{
			BuildingName:   i.BuildingName,
			BuildingNumber: i.BuildingNumber,
			PostTown:       i.PostTown,
			Postcode:       i.Postcode,
			Street:         i.Street,
			SubBuilding:    i.SubBuilding,
		},
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err = enc.Encode(out)

	if err != nil {
		return nil, err
	}

	return buf, nil
}
