package main

import (
	"log"
	"net/http"
	"fmt"
	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/reader"
)

func main() {

    // Set up HTTP server
	http.Handle("/", http.FileServer(http.Dir("./static")))
    http.HandleFunc("/upload", reader.HandleUpload)
    http.HandleFunc("/health", reader.HealthCheckHandler)
    
    // Start HTTP server
    port := 8080
    fmt.Printf("Starting HTTP server on port %d...\n", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}