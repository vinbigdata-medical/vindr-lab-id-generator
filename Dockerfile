FROM golang:1.15.2 AS build

RUN mkdir /opt/app
WORKDIR /opt/app

COPY ./go.mod ./go.mod
RUN go mod download

COPY ./*.go ./
COPY ./generator ./generator
COPY ./constants ./constants
COPY ./utils ./utils

RUN go build -o bin main.go

FROM debian:stable-slim

RUN mkdir /opt/app
WORKDIR  /opt/app

COPY --from=build /opt/app/bin /opt/app/bin
COPY ./conf ./conf

CMD ["./bin"]
