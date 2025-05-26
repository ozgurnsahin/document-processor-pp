package reader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"

	models "github.com/ozgurnsahin/document-processor-pp/document-ingestion/data_models"
	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/processor"
	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/storage"
)


func FileReader(content []byte, filename string, fileSize int64) (*models.Document, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("file content is empty")
	}
	
	mtype := mimetype.Detect(content)
	contentType := strings.TrimSpace(mtype.String())

	if !isSupportedFileType(contentType) {
		return nil, fmt.Errorf("unsupported file type: %s", contentType)
	}

	doc := &models.Document{
		FileName: filename,
		Content: content,
		ContentType: contentType,
		Size: fileSize,
		Status: models.StatusReceived,
	}

	return doc, nil
}

func isSupportedFileType(mimeType string) bool {
	supportedTypes := map[string]bool{
		"application/pdf": 				  true,
		"text/plain; charset=utf-8":      true,
		"text/rtf; charset=utf-8":        true,
	}
	return supportedTypes[mimeType]
}

func HandleUpload(w http.ResponseWriter, r *http.Request, client *processor.Client, mongodb *storage.MongoDB) {
	// Checks if the method is allowed
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
	}

	// Cheks the the file size 
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		http.Error(w, "Error file exceeds file limir" +err.Error(), http.StatusBadRequest)
		return 
	}

	// Reads file 
	file, header, err := r.FormFile("document")
	if err != nil {
        http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

	if header.Size > 5*1024*1024 { // 5MB
        http.Error(w, "File too large (max 5MB)", http.StatusBadRequest)
        return
    }

	fileContent, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	doc, err := FileReader(fileContent, header.Filename, header.Size) 
	if err != nil {
        http.Error(w, "Error reading tempfile: "+err.Error(), http.StatusInternalServerError)
        return
    }

	doc.ID = uuid.New().String()
	doc.Status = models.StatusProcessing
	doc.UploadedAt = time.Now()

	err = mongodb.InsertDocuments(doc)
	if err != nil {
		http.Error(w, "Error saving document: "+err.Error(), http.StatusInternalServerError)
        return
	}

	chunks,err := client.ProcessDocument(doc)
	if err != nil {
		http.Error(w, "Error at communications process: "+err.Error(), http.StatusInternalServerError)
        return
	}

	err = mongodb.InsertChunks(doc.ID, chunks)
    if err != nil {
        http.Error(w, "Error saving chunks: "+err.Error(), http.StatusInternalServerError)
        return
    }

	fmt.Printf("Successfully processed document %s with %d chunks\n", doc.ID, len(chunks))
    
    // Update document status
    doc.Status = models.StatusCompleted
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
        "document_id": doc.ID,
        "filename":    doc.FileName,
        "status":      doc.Status,
        "size":        doc.Size,
    })
    
}

func HandleSearch(w http.ResponseWriter, r *http.Request, client *processor.Client, m *storage.MongoDB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
	}

	var request struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
        return
	}

	if request.Query == "" {
		http.Error(w, "Query can not be empty", http.StatusBadRequest)
        return
	}

	queryVector, err := client.CreateInputEmbeddings(request.Query)
	if err != nil {
		http.Error(w, "Error creating embedding: "+err.Error(), http.StatusInternalServerError)
        return
	}

	documentNames, err := m.SearchDocumetns(queryVector)
	if err != nil {
		http.Error(w, "Search failed: "+err.Error(), http.StatusInternalServerError)
        return
	}

	var response map[string]interface{}
	if len(documentNames) == 0 {
		response = map[string]interface{}{
            "documents": []string{},
            "message":   "No similar documents found",
        }
	} else {
		response = map[string]interface{}{
            "documents": documentNames,
            "message":  "Similar documents returned",
        }
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Service is healthy!")
}

