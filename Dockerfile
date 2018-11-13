FROM golang:1.11.1-alpine AS builder

RUN apk add bash ca-certificates git

WORKDIR /application

ENV GO111MODULE=on
ENV MAXMIND_GEOIP_DB_PATH=/application/etc/maxmind/GeoLite2-City.mmdb

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN swag init -s ./web/static/swagger
RUN CGO_ENABLED=0 GOOS=linux go build -a -o $GOPATH/bin/p1pay_api .

ENTRYPOINT $GOPATH/bin/p1pay_api -migration=up && $GOPATH/bin/p1pay_api