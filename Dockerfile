FROM golang:1.23.6 AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o bin/api ./cmd/api
RUN go build -o bin/ingest ./cmd/ingest
RUN go build -o bin/processing ./cmd/processing

FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /app/bin /app/bin