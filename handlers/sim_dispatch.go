package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/utilitywarehouse/json-rpc-proxy/extpoints"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	simDispatchRoute  = "simdispatch"
	kafkaProducerHost = "http-kafka-producer:80"
	kafkaTopic        = "OutboundBillSimRequestEvents"
)

type SimDispatchRequested struct {
	AccountId          string    `json:"accountId"`
	DestinationAddress string    `json:"destinationAddress"`
	Cli                string    `json:"cli"`
	OldSimNumber       string    `json:"oldSimNumber"`
	IvrFields          IvrFields `json:"ivrFields"`
}

type IvrFields struct {
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

	simDispatchRequested := getSimDispatchRequested(requestBody)

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

func getSimDispatchRequested(requestBody []byte) *SimDispatchRequested {
	requestBodyString := string(requestBody)
	fields := strings.Split(requestBodyString, ",")

	ivrFields := &IvrFields{
		BankAccountLastFourDigits: fields[4],
		MobSecurity:               fields[5],
		DateOfBirth1:              fields[6],
		DateOfBirth2:              fields[7],
	}

	simDispatchRequested := &SimDispatchRequested{
		AccountId:          fields[0],
		DestinationAddress: fields[1],
		Cli:                fields[2],
		OldSimNumber:       fields[3],
		IvrFields:          *ivrFields,
	}

	return simDispatchRequested
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
