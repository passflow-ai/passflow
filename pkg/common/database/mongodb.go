package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	mongoClient   *mongo.Client
	mongoDatabase *mongo.Database
	mongoOnce     sync.Once
	mongoErr      error
)

// Connect establishes a connection to MongoDB.
func Connect(ctx context.Context, uri, dbName string) error {
	mongoOnce.Do(func() {
		clientOptions := options.Client().
			ApplyURI(uri).
			SetMaxPoolSize(100).
			SetMinPoolSize(10).
			SetMaxConnIdleTime(30 * time.Second).
			SetServerSelectionTimeout(10 * time.Second).
			SetConnectTimeout(10 * time.Second)

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			mongoErr = fmt.Errorf("failed to connect to MongoDB: %w", err)
			return
		}

		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			mongoErr = fmt.Errorf("failed to ping MongoDB: %w", err)
			return
		}

		mongoClient = client
		mongoDatabase = client.Database(dbName)

		fmt.Printf("Connected to MongoDB database: %s\n", dbName)
	})

	return mongoErr
}

// GetClient returns the MongoDB client.
func GetClient() *mongo.Client {
	return mongoClient
}

// GetDatabase returns the MongoDB database.
func GetDatabase() *mongo.Database {
	return mongoDatabase
}

// GetCollection returns a MongoDB collection.
func GetCollection(name string) *mongo.Collection {
	if mongoDatabase == nil {
		return nil
	}
	return mongoDatabase.Collection(name)
}

// IsConnected checks if MongoDB is connected.
func IsConnected() bool {
	if mongoClient == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return mongoClient.Ping(ctx, readpref.Primary()) == nil
}

// Disconnect closes the MongoDB connection.
func Disconnect(ctx context.Context) error {
	if mongoClient == nil {
		return nil
	}
	return mongoClient.Disconnect(ctx)
}

// CreateIndexes creates indexes for a collection.
func CreateIndexes(ctx context.Context, collectionName string, indexes []mongo.IndexModel) error {
	collection := GetCollection(collectionName)
	if collection == nil {
		return fmt.Errorf("collection %s not found", collectionName)
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}
