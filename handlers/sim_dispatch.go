package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/utilitywarehouse/json-rpc-proxy/extpoints"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	simDispatchRoute  = "simdispatch"
	kafkaProducerHost = "http-kafka-producer:8080"
	kafkaTopic        = "OutboundBillSimRequestEvents"
)

type SimDispatchRequestedEvent struct {
	AccountId                 string `json:"accountId"`
	Destination               string `json:"destination"`
	CLI                       string `json:"cli"`
	DispatchAt                string `json:"dispatchAt"`
	YearOfBirth               int    `json:"yearOfBirth"`
	BankAccountLastFourDigits string `json:"bankAccountLastFourDigits"`
}

func init() {
	log.Print("registering sim dispatch handler")
	ok := endpoints.Register(extpoints.Endpoint(func() http.HandlerFunc { return handleSimDispatchRequest }), simDispatchRoute)
	if !ok {
		log.Panicf("handler name: %s failed to register", route)
	}
}

func handleSimDispatchRequest(wr http.ResponseWriter, req *http.Request) {
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(wr, fmt.Sprintf("error reading request body %v", err), http.StatusBadRequest)
		return
	}

	simDispatchRequestedEvent, err := getSimDispatchRequestedEvent(requestBody)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	jsonBytes, err := json.Marshal(simDispatchRequestedEvent)
	if err != nil {
		http.Error(wr, fmt.Sprintf("error encoding SimDispatchRequestedEvent %+v", err), http.StatusInternalServerError)
		return
	}

	err = produceSimDispatchRequestedEvent(jsonBytes)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadGateway)
		return
	}
}

func getSimDispatchRequestedEvent(requestBody []byte) (*SimDispatchRequestedEvent, error) {
	requestBodyString := string(requestBody)
	split := strings.Split(requestBodyString, ",")

	yearOfBirth, err := strconv.Atoi(split[4])
	if err != nil {
		return nil, fmt.Errorf("error parsing int from year of birth %v", err)
	}

	simDispatchRequestedEvent := &SimDispatchRequestedEvent{
		AccountId:                 split[0],
		Destination:               split[1],
		CLI:                       split[2],
		DispatchAt:                split[3],
		YearOfBirth:               yearOfBirth,
		BankAccountLastFourDigits: split[5],
	}

	return simDispatchRequestedEvent, err
}

func produceSimDispatchRequestedEvent(payload []byte) error {
	httpClient := http.Client{}
	producerResponse, err := httpClient.Post(
		fmt.Sprintf("http://%s/produce/%s", kafkaProducerHost, kafkaTopic),
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("error getting response from kafka producer %v", err)
	}
	if producerResponse.StatusCode != 200 {
		return fmt.Errorf(fmt.Sprintf("received non 200 response from kafka producer: %v, %v", producerResponse.StatusCode, producerResponse.Status))
	}
	return nil
}
