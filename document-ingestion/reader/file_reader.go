package reader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"

	models "github.com/ozgurnsahin/document-processor-pp/document-ingestion/data_models"
	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/processor"
	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/storage"
)


func FileReader(filePath string) (*models.Document, error) {

	if _, err := os.Stat(filePath); os.IsNotExist(err){
		return nil, fmt.Errorf("file does not exists: %s", filePath)
	}

	content, err := os.ReadFile(filePath) 

	if err != nil {
		return nil, fmt.Errorf("error occured while file reading")
	}

	fileInfo, errinfo := os.Stat(filePath) 
	
	if errinfo != nil {
		return nil, fmt.Errorf("file does not exists: %s", filePath)
	}
	
	mtype := mimetype.Detect(content)
	contentType := strings.TrimSpace(mtype.String())

	if !isSupportedFileType(contentType) {
		return nil, fmt.Errorf("unsupported file type: %s", contentType)
	}

	doc := &models.Document{
		FileName: filepath.Base(filePath),
		Content: content,
		ContentType: contentType,
		Size: fileInfo.Size(),
		Status: models.StatusReceived,
	}

	return doc, nil
}

func isSupportedFileType(mimeType string) bool {
	supportedTypes := map[string]bool{
		"application/pdf": true,
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

	// Create a temperory files 
	tempFile, err := os.CreateTemp("./upload", "upload-*.txt")
	if err != nil {
        http.Error(w, "Error creating tempfile: "+err.Error(), http.StatusInternalServerError)
        return
    }

	_, err = io.Copy(tempFile, file)
	if err != nil {
        http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
        return
    }

	doc, err := FileReader(tempFile.Name()) 
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

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Service is healthy!")
}

