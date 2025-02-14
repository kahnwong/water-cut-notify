FROM golang:1.24-alpine AS build-stage

WORKDIR /app

# https://github.com/uber/h3/issues/354
RUN apk add --no-cache build-base

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=1 go build -ldflags "-w -s" -o /water-cut-notify

# hadolint ignore=DL3007
FROM alpine:latest AS build-release-stage

WORKDIR /

COPY --from=build-stage /water-cut-notify /water-cut-notify

RUN chmod +x /water-cut-notify

ENTRYPOINT ["/water-cut-notify"]
