# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine AS builder
ENV GOWORK=off

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
COPY go-ios ./go-ios 
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -o server .


FROM alpine:3.22.2
ENV GOWORK=off

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
ENTRYPOINT ["./server"]