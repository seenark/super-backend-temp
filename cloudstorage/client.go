package cloudstorage

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"time"

	"bitbucket.org/atiwataqs/super-backend/config"
	"cloud.google.com/go/storage"
)

type ClientUploader struct {
	cl         *storage.Client
	projectID  string
	bucketName string
	uploadPath string
}

func NewGoogleStorageUploader() *ClientUploader {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "./aumaum-can-dlt-on-iam-setting-1a8fa2f46228.json") // FILL IN WITH YOUR FILE PATH
	cfg := config.GetConfig()
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	uploader := &ClientUploader{
		cl:         client,
		bucketName: cfg.Google.BucketName,
		projectID:  cfg.Google.ProjectID,
		uploadPath: "superx-test/",
	}

	return uploader
}

// UploadFile uploads an object
func (c *ClientUploader) UploadFile(file multipart.File, object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := c.cl.Bucket(c.bucketName).Object(c.uploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

func (c *ClientUploader) DeleteFile(fileName string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	err := c.cl.Bucket(c.bucketName).Object(c.uploadPath + fileName).Delete(ctx)
	if err != nil {
		return fmt.Errorf("Delete: %v", err)
	}
	return nil

}
