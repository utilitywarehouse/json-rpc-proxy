package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/utilitywarehouse/json-rpc-proxy/extpoints"
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
	Address          broadbandAvailabilityRequestAddress `json:"address"`
	Cli              string                              `json:"cli"`
	ProductRequested string                              `json:"productRequested"`
	Source           string                              `json:"source"`
}

type broadbandAvailabilityRequestAddress struct {
	BuildingName   string `json:"buildingName"`
	BuildingNumber string `json:"buildingNumber"`
	PostTown       string `json:"postTown"`
	Postcode       string `json:"postcode"`
	Street         string `json:"street"`
	SubBuilding    string `json:"subBuilding"`
}

const route = "getbroadbandavailability/max"

var (
	broadbandAvailabilityURL = "https://linechecker-telecom.%s.uw.systems/api/broadbandavailability/max"
	endpoints                = extpoints.Endpoints
	env                      = envOrPanic()
)

func envOrPanic() string {
	env, ok := os.LookupEnv("env")
	if !ok {
		log.Panic("Could not find env in environment variables")
	}
	return env
}

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
	sr, err := cl.Post(fmt.Sprintf(broadbandAvailabilityURL, env), "application/json", buf)
	if err != nil {
		http.Error(wr, fmt.Sprintf("error getting response from upstream service %+v", err), http.StatusBadGateway)
	}

	wr.Header().Set("Content-Type", "application/json")

	_, err = io.Copy(wr, sr.Body)
	if err != nil {
		http.Error(wr, fmt.Sprintf("error writing to response body: %+v", err), http.StatusInternalServerError)
	}
}

func getServiceRequest(body io.ReadCloser) (io.Reader, error) {
	dec := json.NewDecoder(body)
	incomingReq := &incomingBroabdandAvailabilityRequest{}
	err := dec.Decode(incomingReq)
	if err != nil {
		return nil, err
	}

	out := &broadbandAvailabilityRequest{
		Cli:              incomingReq.Cli,
		ProductRequested: incomingReq.ProductRequested,
		Source:           incomingReq.Source,
		Address: broadbandAvailabilityRequestAddress{
			BuildingName:   incomingReq.BuildingName,
			BuildingNumber: incomingReq.BuildingNumber,
			PostTown:       incomingReq.PostTown,
			Postcode:       incomingReq.Postcode,
			Street:         incomingReq.Street,
			SubBuilding:    incomingReq.SubBuilding,
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
