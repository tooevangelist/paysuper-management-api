FROM golang:1.11-alpine AS builder

RUN apk add bash ca-certificates git

WORKDIR /application

ENV GO111MODULE=on
ENV PSP_ACCOUNTING_CURRENCY=EUR
ENV PATH_TO_PS_CONFIG=/application/config/parameters/parameters.yml
ENV AMQP_ADDRESS=amqp://p1pay-rabbitmq
ENV CENTRIFUGO_SECRET=3BHvbrHkThYJ6J8Knd4DCsbL

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -o $GOPATH/bin/paysuper_management_api .

ENTRYPOINT $GOPATH/bin/paysuper_management_api -migration=up && $GOPATH/bin/paysuper_management_api
