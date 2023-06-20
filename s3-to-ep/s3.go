package main

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

func fetchS3Content(ctx context.Context, s3Client S3Client, record events.S3EventRecord) ([]byte, error) {
	s3Object, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(record.S3.Bucket.Name),
		Key:    aws.String(record.S3.Object.Key),
	})

	if err != nil {
		log.Printf("error fetching s3 object: %s", err)
		return nil, err
	}

	return io.ReadAll(s3Object.Body)
}
