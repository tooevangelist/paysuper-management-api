FROM golang:1.11.1-alpine AS builder

RUN apk add bash ca-certificates git

WORKDIR /application

ENV GO111MODULE=on
ENV MAXMIND_GEOIP_DB_PATH=/application/etc/maxmind/GeoLite2-City.mmdb
ENV PSP_ACCOUNTING_CURRENCY=EUR
ENV PATH_TO_PS_CONFIG=/application/config/parameters/parameters.yml
ENV MICRO_REGISTRY_ADDRESS=p1pay-consul
ENV MICRO_BROKER=rabbitmq
ENV MICRO_BROKER_ADDRESS=amqp://p1pay-rabbitmq
ENV CENTRIFUGO_SECRET=3BHvbrHkThYJ6J8Knd4DCsbL

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -o $GOPATH/bin/p1pay_api .

ENTRYPOINT $GOPATH/bin/p1pay_api -migration=up && $GOPATH/bin/p1pay_api