package storage

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/lgr"
)

type s3Service struct {
	ConfigSvc config.IService
	Client    *s3.Client
}

func newS3(cfgsvc config.IService) IService {
	s := &s3Service{
		ConfigSvc: cfgsvc,
	}
	err := s.makeS3Client(context.Background())
	if err != nil {
		panic(err)
	}

	return s
}

func (svc *s3Service) NewFile(ctx context.Context, folder, filePath, identifier string) (string, error) {
	bucketName := svc.ConfigSvc.GetStorageBucket()
	keyName := fmt.Sprintf("%s/%s", folder, identifier)
	lgr.Logger.Debug("S3.NewFile",
		slog.String("folder", folder),
		slog.String("filePath", filePath),
		slog.String("identifier", identifier),
		slog.String("bucket", bucketName),
		slog.String("key", keyName),
	)

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// WARNING: if the file already exists in S3, it will be overwritten
	_, err = svc.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	// Close the file before deleting it
	file.Close()

	// Delete the local file
	err = os.Remove(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to delete local file: %w", err)
	}

	// URL must be generated from the bucket name and the folder
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, svc.ConfigSvc.GetStorageRegion(), keyName), nil
}

func (svc *s3Service) makeS3Client(ctx context.Context) error {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(svc.ConfigSvc.GetStorageRegion()),
	)

	if err != nil {
		return err
	}

	svc.Client = s3.NewFromConfig(cfg)
	return nil
}
