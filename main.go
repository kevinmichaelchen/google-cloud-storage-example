package main

import (
	"cloud.google.com/go/storage"
	"context"
	"log"
	"os"
)

func main() {
	projectID := os.Getenv("DATASTORE_PROJECT_ID")
	if projectID == "" {
		log.Fatal(`You need to set the environment variable "DATASTORE_PROJECT_ID"`)
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Could not create storage client: %v", err)
	}

	err = createWithAttrs(client, projectID, "glass-of-water")
	if err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}
}

func createWithAttrs(client *storage.Client, projectID, bucketName string) error {
	ctx := context.Background()
	bucket := client.Bucket(bucketName)
	if err := bucket.Create(ctx, projectID, &storage.BucketAttrs{
		Location:     "US",
	}); err != nil {
		return err
	}
	return nil
}
