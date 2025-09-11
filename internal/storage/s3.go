package storage

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	S3Client   *s3.Client
	BucketName string
)

func InitS3() {
	BucketName = os.Getenv("AWS_BUCKET_NAME")
	if BucketName == "" {
		log.Fatal("AWS_BUCKET_NAME environment variable not set")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	S3Client = s3.NewFromConfig(cfg)
}
