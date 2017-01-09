FROM alpine

RUN apk add --no-cache ca-certificates

RUN mkdir /app

WORKDIR /app

ADD json-rpc-proxy json-rpc-proxy

CMD ./json-rpc-proxy
