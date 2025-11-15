# syntax=docker/dockerfile:1.7


FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine AS builder
ENV GOWORK=off
# These are automatically set by BuildKit when you use --platform
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum go-ios ./
RUN go mod download

COPY . .

# Build *for the target platform*, not the build platform
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -o server .

########################
# 2) Run stage
########################
# Use a base image that has variants for all platforms you care about
FROM alpine:3.22.2
ENV GOWORK=off

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
ENTRYPOINT ["./server"]

