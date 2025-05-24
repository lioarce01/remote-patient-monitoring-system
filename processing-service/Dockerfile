FROM golang:1.23 AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY pkg/common ./pkg/common
COPY processing-service ./

RUN go mod download
RUN go build -o bin/ ./cmd

FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /app/bin /app/bin
