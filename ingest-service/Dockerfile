FROM golang:1.23 AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app/ingest-service

COPY pkg/common /app/pkg/common
COPY ingest-service /app/ingest-service

RUN go mod download
RUN go build -o ../bin/ingest-service-binary ./cmd

FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /app/bin/ingest-service-binary /app/bin/ingest-service-binary

CMD ["/app/bin/ingest-service-binary"]
