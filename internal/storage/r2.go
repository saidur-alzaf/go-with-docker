package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"go-sqlite-api/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2StorageService struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

func NewR2StorageService(ctx context.Context, cfg config.R2Config) (*R2StorageService, error) {
	if cfg.AccountID == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.BucketName == "" {
		return nil, fmt.Errorf("R2 configuration is incomplete")
	}

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID),
		}, nil
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithEndpointResolverWithOptions(r2Resolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS R2 config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)

	publicURL := strings.TrimSuffix(cfg.PublicURL, "/")

	return &R2StorageService{
		client:     client,
		bucketName: cfg.BucketName,
		publicURL:  publicURL,
	}, nil
}

func (s *R2StorageService) UploadFile(ctx context.Context, fileReader io.Reader, filename string, contentType string) (string, error) {
	ext := path.Ext(filename)
	key := fmt.Sprintf("products/%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filename, ext), ext)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        fileReader,
		ContentType: aws.String(contentType),
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload object to R2: %w", err)
	}

	var fileURL string
	if s.publicURL != "" {
		fileURL = fmt.Sprintf("%s/%s", s.publicURL, key)
	} else {
		fileURL = fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", s.bucketName, key)
	}

	return fileURL, nil
}
