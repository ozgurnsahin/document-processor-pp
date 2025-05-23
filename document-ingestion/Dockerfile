FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app

RUN mkdir -p /app/static

COPY --from=builder /app/app .
COPY static/ ./static/

EXPOSE 8080
CMD ["./app"]