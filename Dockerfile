# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder

WORKDIR /build

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Transfer source code
COPY *.go ./
COPY conditions.json ./
COPY templates ./templates
COPY src ./src

# Build
RUN CGO_ENABLED=0 go build -trimpath -o /dist/rainbbit

# Test
FROM build-stage AS run-test-stage
RUN go test -v ./...

FROM scratch AS build-release-stage

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY conditions.json /app/
COPY templates /app/templates
COPY --from=builder /dist /app

ENTRYPOINT ["/app/rainbbit"]
