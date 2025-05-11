package models

type Document struct {
	Path string
	Name string
	Content []byte
	Size int64 
}