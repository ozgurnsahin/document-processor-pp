FROM golang:1.19-alpine AS builder

WORKDIR /app

# Pre-copy/cache go.mod for pre-downloading dependencies
COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o document-ingestion main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/document-ingestion .

EXPOSE 8080
EXPOSE 50051

CMD ["./document-ingestion"]