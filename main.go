package main

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type JsonCredentials struct {
	TypeString              string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURI string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

func getCredentials() JsonCredentials {
	path, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !ok {
		log.Fatalf("Environment variable GOOGLE_APPLICATION_CREDENTIALS must be path to JSON credentials")
	}
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("No file found at: %s", path)
	}
	defer f.Close()
	byteValue, _ := ioutil.ReadAll(f)

	var result JsonCredentials
	err = json.Unmarshal([]byte(byteValue), &result)
	if err != nil {
		log.Fatalf("Could not marshal file: %s", path)
	}
	return result
}

func main() {
	credentials := getCredentials()
	projectID := credentials.ProjectID
	serviceAccountEmailAddress := credentials.ClientEmail
	if projectID == "" {
		log.Fatalf("File does not contain project_id")
	}
	if serviceAccountEmailAddress == "" {
		log.Fatalf("File does not contain client_email")
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Could not create storage client: %v", err)
	}

	bucketName := "test-organizations"
	bh := client.Bucket(bucketName)

	// GET OR CREATE BUCKET
	if !bucketExists(bh) {
		log.Printf("Bucket %s does not exist. Creating now...", bucketName)
		err = createWithAttrs(client, projectID, bucketName)
		if err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		} else {
			log.Printf("Successfully created bucket: %s.", bucketName)
		}
	}

	ctx = context.Background()
	r, f := readerFromFile("./dog.txt")
	defer f.Close()
	_, _, err = upload(bh, ctx, r, "iris/dog.txt", true)
	if err != nil {
		log.Fatalf("Failed to upload to bucket: %v", err)
	}

	storage.SignedURL(bucketName, "iris/dog.txt", &storage.SignedURLOptions{
		GoogleAccessID: serviceAccountEmailAddress,
		Method:         http.MethodPut,
		Expires:        time.Now().Add(2 * time.Minute),
	})
}

func bucketExists(bh *storage.BucketHandle) bool {
	// Next check if the bucket exists
	ctx := context.Background()
	_, err := bh.Attrs(ctx)
	exists := err == nil
	return exists
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
		Location: "US",
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
