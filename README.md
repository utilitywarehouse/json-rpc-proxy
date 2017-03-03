# JSON RPC Proxy

This application sits between an application and various internal and external
services.

Requests are made to this service from the client application and in turn this
service executes requests on the clients' behalf. It performs flattening of
response objects to make data easier to manipulate and handle applications that
have trouble with non flat JSON structures.

## Godep

please note that packages are vendored with godep, so if you update any
dependencies (most likely uw services) then you should update the dependencies
with `go get -u ./dependency` followed by `godep save`

## Installation `go get github.com/utilitywarehouse/json-rpc-proxy`
  
## Build

`go test .`
 
`go build .`
