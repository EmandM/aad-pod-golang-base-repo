FROM golang:alpine as builder

RUN apk update \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/* \
    && update-ca-certificates \
    && apk add git

ADD ./src /go/src/test/keyvault-sample
WORKDIR /go/src/test/keyvault-sample

RUN go get -d .
RUN go build . -t keyvault-sample-image

FROM alpine
RUN apk update \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/* \
    && update-ca-certificates

CMD ["go", "run", "main.go"]
