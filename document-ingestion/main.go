package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/processor"
	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/reader"
	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/storage"
)



func main() {

    processorClient, err := processor.NewClient()
    if err != nil {
		log.Fatalf("Failed to initialize processor client: %v", err)
	}
	defer processorClient.Close()

    mongodb, err := storage.NewMongoClient()
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    defer mongodb.Close()
    
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/" {
            http.ServeFile(w, r, "./static/static.html")
            return
        }
        http.FileServer(http.Dir("./static")).ServeHTTP(w, r)
    })

    http.HandleFunc("/upload",func(w http.ResponseWriter, r *http.Request){
        reader.HandleUpload(w, r, processorClient, mongodb)
    })
    http.HandleFunc("/health", reader.HealthCheckHandler)

    // Start HTTP server
    port := 8080
    fmt.Printf("Starting HTTP server on port %d...\n", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}