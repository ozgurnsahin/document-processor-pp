package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/processor"
	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/reader"
)



func main() {

    ServiceAddr := os.Getenv("PROCESSING_SERVICE_ADDR")
    if ServiceAddr == "" {
        ServiceAddr = "document-processing:50052"
    }

    processorClient, err := processor.NewClient(ServiceAddr)
    if err != nil {
		log.Fatalf("Failed to initialize processor client: %v", err)
	}
	defer processorClient.Close()

    // Set up HTTP server
	http.Handle("/", http.FileServer(http.Dir("./static")))
    http.HandleFunc("/upload",func(w http.ResponseWriter, r *http.Request){
        reader.HandleUpload(w, r, processorClient)
    })
    http.HandleFunc("/health", reader.HealthCheckHandler)
    
    // Start HTTP server
    port := 8080
    fmt.Printf("Starting HTTP server on port %d...\n", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}