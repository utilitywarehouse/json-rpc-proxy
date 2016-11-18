package extpoints

import "net/http"

type Endpoint func() http.HandlerFunc
