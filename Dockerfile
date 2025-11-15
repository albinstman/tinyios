# Pick which go-ios version/commit/tag to build (defaults to main)
ARG GO_IOS_VERSION=v1.0.182

# --- builder ---
FROM golang:1.22 AS builder
ARG GO_IOS_VERSION
WORKDIR /src

# Checkout the exact version
RUN git clone --depth 1 --branch ${GO_IOS_VERSION} https://github.com/danielpaulus/go-ios.git .

# Build the REST API after generating Swagger docs (creates ./restapi/docs)
WORKDIR /src/restapi
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3
RUN /go/bin/swag init --parseDependency
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/go-ios-restapi .

# --- runtime ---
FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/go-ios-restapi /usr/local/bin/go-ios-restapi
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/go-ios-restapi"]