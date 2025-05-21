package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct{
	client    *mongo.Client
	database  *mongo.Database
	documents *mongo.Collection
	chunks	  *mongo.Collection	
}


func NewMongoClient(dbName string) (*MongoDB, error) {
	mongoURI := os.Getenv("MONGODB_STRING")

	clientInfos := options.Client().ApplyURI(mongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientInfos)
	if err != nil {
        return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
    }

	err = client.Ping(ctx, nil)
	if err != nil {
        return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
    }

	db := client.Database(dbName)
	documents := db.Collection("documents")
	chunks := db.Collection("chunks")

	_, err = documents.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{bson.E{Key: "id", Value: 1}},
        Options: options.Index().SetUnique(true),
    })
    if err != nil {
        log.Printf("Warning: Failed to create document index: %v", err)
    }
    
    _, err = chunks.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{bson.E{Key: "document_id", Value: 1}},
    })
    if err != nil {
        log.Printf("Warning: Failed to create chunks index: %v", err)
    }
    
    log.Printf("Connected to MongoDB: %s", mongoURI)
    
    return &MongoDB{
        client:    client,
        database:  db,
        documents: documents,
        chunks:    chunks,
    }, nil
}
