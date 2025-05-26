package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	models "github.com/ozgurnsahin/document-processor-pp/document-ingestion/data_models"
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


func NewMongoClient() (*MongoDB, error) {
	err := godotenv.Load()
    if err != nil {
        log.Printf("Warning: Error loading .env file: %v", err)
    }

	mongoURI := os.Getenv("MONGODB_STRING")
	dbName := os.Getenv("MONGODB_DB")
    if dbName == "" {
        dbName = "docDev"
    }
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

func (m *MongoDB) InsertDocuments(doc *models.Document) error{
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 300)
	defer cancel()

	bsonDoc := bson.M{
		"id":           doc.ID,
        "filename":     doc.FileName,
        "content_type": doc.ContentType,
        "size":         doc.Size,
        "uploaded_at":  doc.UploadedAt,
        "status":       doc.Status,
	}

	opt := options.Update().SetUpsert(true)
	_, err := m.documents.UpdateOne(
		ctx,
		bson.M{"id": doc.ID},
		bson.M{"$set": bsonDoc},
		opt,
	)

	if err != nil {
		return fmt.Errorf("failed to save document: %w", err)
	}

	return nil
}

func (m *MongoDB) InsertChunks(documentID string, chunks []*models.DocumentChunk) error{
	if len(chunks) == 0 {
        return nil
    }

	ctx, cancel := context.WithTimeout(context.Background(), 300* time.Second)
	defer cancel()

	_, err := m.chunks.DeleteMany(ctx, bson.M{"document_id": documentID})

	if err != nil {
		return fmt.Errorf("failed to delete existing chunks: %w", err)
	}

	var chunksToInsert []interface{}
    for _, chunk := range chunks {
        chunksToInsert = append(chunksToInsert, bson.M{
            "document_id": chunk.DocumentID,
            "chunk_index": chunk.ChunkIndex,
            "text":        chunk.Text,
            "vector":      chunk.Vector,
        })
    }

	_, err = m.chunks.InsertMany(ctx, chunksToInsert)
    if err != nil {
        return fmt.Errorf("failed to insert chunks: %w", err)
    }

	return nil
}

func (m *MongoDB) SearchDocumetns(queryVector []float32) ([]string, error){
	context, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	pipeline := bson.A{
		bson.M{
			"$vectorSearch": bson.M{
				"index":         "vector_index",
				"path":          "vector", 
				"queryVector":   queryVector,
				"numCandidates": 100,
				"limit":         5,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"score" : bson.M{
					"$meta": "vectorSearchScore"},
				},
			},
		bson.M{
			"$match": bson.M{
				"score": bson.M{"$gte": 0.6},
				},
			},
		bson.M{
			"$project": bson.M{
				"document_id": 1,
				"score": 1,
				},
			},
	}

	cursor, err := m.chunks.Aggregate(context, pipeline)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	defer cursor.Close(context)

	documentIDs := make(map[string]bool)
	for cursor.Next(context) {
		var result struct {
			DocumentID string  `bson:"document_id"`
            Score      float64 `bson:"score"`
		}

		if err := cursor.Decode(&result); err != nil {
            continue
        }

		documentIDs[result.DocumentID] = true
	}

	if len(documentIDs) == 0 {
		return []string{},nil
	}

	return m.getDocuments(documentIDs)

}

func (m *MongoDB) getDocuments(documentids map[string]bool) ([]string, error){
	context, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	ids := make([]string ,0 ,len(documentids))
	for id := range documentids {
		ids = append(ids, id)
	}

	cursor, err := m.documents.Find(context,  bson.M{"id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("failed to get document names: %w", err)
	}
	defer cursor.Close(context)

	var documents []models.Document
	if err := cursor.All(context, &documents); err != nil{
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	names := make([]string,len(documents))
	for i, doc := range documents {
		names[i] = doc.FileName
	}

	return names, nil
}

func (m *MongoDB) Close() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := m.client.Disconnect(ctx); err != nil {
        return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
    }
    
    return nil
}