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


const (
	StatusReceived   = "received"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)