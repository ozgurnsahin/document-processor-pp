package processor

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	models "github.com/ozgurnsahin/document-processor-pp/document-ingestion/data_models"
	pb "github.com/ozgurnsahin/document-processor-pp/document-ingestion/proto"
)

type Client struct{
  	serviceAddr string
	client      pb.DocumentProcessorServiceClient
	conn        *grpc.ClientConn
}

func NewClient(serviceAddr string) (*Client, error){
	conn, err := grpc.NewClient(serviceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil{
		return nil, fmt.Errorf("connection with the grpc server could not created: %w", err)
	}

	client := pb.NewDocumentProcessorServiceClient(conn)

	return &Client{
		serviceAddr: serviceAddr,
		client: client,
		conn: conn,
	},nil

}

func (c *Client) ProcessDocument(doc *models.Document) error{
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 30)
	defer cancel()

	req := &pb.ProcessRequest{
		DocumentId: doc.ID,
		Filename: doc.FileName,
		Content: doc.Content,
		ContentType: doc.ContentType,
	}

	_, err := c.client.ProcessDocument(ctx, req)
	if err != nil {
		return fmt.Errorf("error calling processing service: %w", err)
	}

	fmt.Printf("Document sent for processing with ID: %s\n", doc.ID)
	return nil

}