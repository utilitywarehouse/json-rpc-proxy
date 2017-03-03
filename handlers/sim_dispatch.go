package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/utilitywarehouse/json-rpc-proxy/extpoints"
	"github.com/utilitywarehouse/sim-dispatch-api/events"
)

const (
	simDispatchRoute  = "simdispatch"
	kafkaProducerHost = "http-kafka-producer:80"
	kafkaTopic        = "OutboundBillSimRequestEvents"
)

type IncomingSimDispatchRequest struct {
	AccountId                 string `json:"accountId"`
	DestinationAddress        string `json:"destinationAddress"`
	Cli                       string `json:"cli"`
	OldSimNumber              string `json:"oldSimNumber"`
	BankAccountLastFourDigits string `json:"bankAccountLastFourDigits"`
	MobSecurity               string `json:"mobSecurity"`
	DateOfBirth1              string `json:"dateOfBirth1"`
	DateOfBirth2              string `json:"dateOfBirth2"`
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

	simDispatchRequested, err := getSimDispatchRequested(requestBody)
	if err != nil {
		http.Error(wr, fmt.Sprintf("error generating SimDispatchRequestedEvent %v", err), http.StatusBadRequest)
		return
	}

	jsonBytes, err := json.Marshal(simDispatchRequested)
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

func getSimDispatchRequested(requestBody []byte) (*events.SimDispatchRequested, error) {
	incomingSimDispatchRequest := &IncomingSimDispatchRequest{}
	err := json.Unmarshal(requestBody, incomingSimDispatchRequest)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling IncomingSimDispatchRequest %v", err)
	}

	ivrFields := &events.IvrFields{
		BankAccountLastFourDigits: incomingSimDispatchRequest.BankAccountLastFourDigits,
		MobSecurity:               incomingSimDispatchRequest.MobSecurity,
		DateOfBirth1:              incomingSimDispatchRequest.DateOfBirth1,
		DateOfBirth2:              incomingSimDispatchRequest.DateOfBirth2,
	}

	simDispatchRequested := &events.SimDispatchRequested{
		AccountId:          incomingSimDispatchRequest.AccountId,
		DestinationAddress: incomingSimDispatchRequest.DestinationAddress,
		Cli:                incomingSimDispatchRequest.Cli,
		OldSimNumber:       incomingSimDispatchRequest.OldSimNumber,
		IvrFields:          *ivrFields,
	}

	return simDispatchRequested, nil
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
