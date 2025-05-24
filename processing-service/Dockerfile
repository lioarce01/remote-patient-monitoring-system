FROM golang:1.23 AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app/processing-service

COPY pkg/common /app/pkg/common
COPY processing-service /app/processing-service

RUN go mod download
RUN go build -o ../bin/processing-service-binary ./cmd

FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /app/bin/processing-service-binary /app/bin/processing-service-binary

CMD ["/app/bin/processing-service-binary"]
