FROM golang:1.23 AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app/api-service

COPY pkg/common /app/pkg/common
COPY api-service /app/api-service

RUN go mod download
RUN go build -o ../bin/api-service-binary ./cmd

FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /app/bin/api-service-binary /app/bin/api-service-binary

CMD ["/app/bin/api-service-binary"]
