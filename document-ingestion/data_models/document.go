package models

import (
	"time"
)

type Document struct {
	ID string 			  `json:"id"`
	FileName string		  `json:"filename"`
	ContentType string	  `json:"content_type"`
	Content []byte		  `json:"-"`
	Size int64 			  `json:"size"`
	UploadedAt  time.Time `json:"uploaded_at"`
	Status      string    `json:"status"`
}

type DocumentChunk struct {
    DocumentID  string    `json:"document_id" bson:"document_id"`
    ChunkIndex  int       `json:"chunk_index" bson:"chunk_index"`
    Text        string    `json:"text" bson:"text"`
    Vector      []float32 `json:"vector" bson:"vector"`
}

const (
	StatusReceived   = "received"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
)