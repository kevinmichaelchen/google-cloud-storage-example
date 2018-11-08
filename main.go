package main

import (
	"cloud.google.com/go/storage"
	"context"
	"io"
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

	bucketName := "glass-of-water"

	// CREATE BUCKET
	//err = createWithAttrs(client, projectID, bucketName)
	//if err != nil {
	//	log.Fatalf("Failed to create bucket: %v", err)
	//}

	ctx = context.Background()
	r, f := readerFromFile("./dog.txt")
	defer f.Close()
	bh := client.Bucket(bucketName)
	//client.Buckets()
	_, _, err = upload(bh, ctx, r, "iris/dog.txt", true)
	if err != nil {
		log.Fatalf("Failed to upload to bucket: %v", err)
	}
}

func readerFromFile(path string) (io.Reader, *os.File) {
	var r io.Reader
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	r = f
	return r, f
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

func upload(bh *storage.BucketHandle, ctx context.Context, r io.Reader, name string, public bool) (*storage.ObjectHandle, *storage.ObjectAttrs, error) {
	// Next check if the bucket exists
	if _, err := bh.Attrs(ctx); err != nil {
		return nil, nil, err
	}

	// name must consist entirely of valid UTF-8-encoded runes
	obj := bh.Object(name)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, r); err != nil {
		return nil, nil, err
	}
	if err := w.Close(); err != nil {
		return nil, nil, err
	}

	if public {
		if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
			return nil, nil, err
		}
	}

	attrs, err := obj.Attrs(ctx)
	return obj, attrs, err
}