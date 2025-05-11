package main

import (
	"log"

	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/reader"
)

func main() {
	
	docpath := "sample-documents/sample.txt"

	doc, err := reader.FileReader(docpath)

	if err != nil {
		log.Fatalf("File cant be readed: %s\n", docpath)
	}

	log.Printf("Document content: %s \n", doc.Content)

}