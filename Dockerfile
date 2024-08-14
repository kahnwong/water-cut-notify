FROM golang:1.23-alpine AS build-stage

WORKDIR /app

# https://github.com/uber/h3/issues/354
RUN apk add build-base

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=1 go build -ldflags "-w -s" -o /water-cut-notify

FROM alpine:latest AS build-release-stage

WORKDIR /

COPY --from=build-stage /water-cut-notify /water-cut-notify

RUN chmod +x /water-cut-notify

ENTRYPOINT ["/water-cut-notify"]
