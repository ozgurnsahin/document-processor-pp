package processor

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
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

func NewClient() (*Client, error){
	err := godotenv.Load()
    if err != nil {
        log.Printf("Warning: Error loading .env file: %v", err)
    }

	serviceAddr := os.Getenv("PROCESSING_SERVICE_ADDR")

	maxMsgSize := 10 * 1024 * 1024 // 10MB

	conn, err := grpc.NewClient(serviceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), 
			grpc.WithDefaultCallOptions(
            grpc.MaxCallRecvMsgSize(maxMsgSize),
            grpc.MaxCallSendMsgSize(maxMsgSize),
        ))
	if err != nil{
		return nil, fmt.Errorf("connection with the grpc server could not created: %w", err)
	}

	client := pb.NewDocumentProcessorServiceClient(conn)

	return &Client{
		serviceAddr: serviceAddr,
		client: client,
		conn: conn,
	}, nil

}

func (c *Client) ProcessDocument(doc *models.Document) ([]*models.DocumentChunk, error){
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 600)
	defer cancel()

	req := &pb.ProcessRequest{
		DocumentId: doc.ID,
		Filename: doc.FileName,
		Content: doc.Content,
		ContentType: doc.ContentType,
	}

	resp, err := c.client.ProcessDocument(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error calling processing service: %w", err)
	}

	fmt.Printf("Document sent for processing with ID: %s\n", doc.ID)

	if resp.Status != "completed" {
		return nil, fmt.Errorf("processing failed: %s", resp.Error)
	}

	chunks := make([]*models.DocumentChunk, 0, len(resp.Chunks))
    for i, chunk := range resp.Chunks {
        chunks = append(chunks, &models.DocumentChunk{
            DocumentID: doc.ID,
            ChunkIndex: i,
            Text:       chunk.Text,
            Vector:     chunk.Vector,
        })
    }

    fmt.Printf("Received %d processed chunks for document: %s\n", len(chunks), doc.ID)

    return chunks, nil

}

func (c *Client) Close() error{
	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}