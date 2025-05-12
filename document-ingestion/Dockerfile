FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

RUN mkdir -p /app/uploads
RUN mkdir -p /app/sample-documents
RUN mkdir -p /app/static

RUN echo "This is a sample document." > /app/sample-documents/sample.txt

EXPOSE 8080

CMD ["/app/app"]