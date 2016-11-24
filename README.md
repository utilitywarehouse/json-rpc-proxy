# UW Bill RPC Handler

This application sits between Bill and various internal and external services.

Requests are made to this service from Bill and in turn this service executes requests 
on Bill's behalf. It performs flattening of response objects to make data
 easier to manipulate and handle in Equinox.
  
  
## Build

`go test .`
 
`go build .`
