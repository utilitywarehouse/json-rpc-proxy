FROM alpine

RUN apk add --no-cache ca-certificates

RUN mkdir /app

WORKDIR /app

ADD uw-bill-rpc-handler uw-bill-rpc-handler

CMD uw-bill-rpc-handler
