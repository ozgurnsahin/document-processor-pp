package reader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	models "github.com/ozgurnsahin/document-processor-pp/document-ingestion/data_models"
	"github.com/ozgurnsahin/document-processor-pp/document-ingestion/processor"
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
	

	doc := &models.Document{
		FileName: filepath.Base(filePath),
		Content: content,
		ContentType: filepath.Base(filePath),
		Size: fileInfo.Size(),
		Status: models.StatusReceived,
	}

	return doc, nil
}

func HandleUpload(w http.ResponseWriter, r *http.Request, client *processor.Client) {
	// Checks if the method is allowed
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
	}

	// Cheks the the file size 
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Error file exceeds file limir" +err.Error(), http.StatusBadRequest)
		return 
	}

	// Reads file 
	file, _, err := r.FormFile("document")
	if err != nil {
        http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

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

	err = client.ProcessDocument(doc)
	if err != nil {
		http.Error(w, "Error at communications process: "+err.Error(), http.StatusInternalServerError)
        return
	}

}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Service is healthy!")
}

