package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
)

type staticTestS3Client struct {
	params *s3.GetObjectInput
	output *s3.GetObjectOutput
	err    error
}

func (s *staticTestS3Client) GetObject(_ context.Context, params *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	s.params = params
	return s.output, s.err
}

func Test_fetchS3Content_success(t *testing.T) {
	const (
		bucketName = "test-bucket"
		s3Content  = "test-content"
		s3Key      = "test-key"
	)
	s3Client := &staticTestS3Client{
		output: &s3.GetObjectOutput{
			Body: io.NopCloser(strings.NewReader(s3Content)),
		},
	}
	record := events.S3EventRecord{
		S3: events.S3Entity{
			Bucket: events.S3Bucket{
				Name: bucketName,
			},
			Object: events.S3Object{
				Key: s3Key,
			},
		},
	}
	content, err := fetchS3Content(context.Background(), s3Client, record)
	assert.NoError(t, err)
	assert.Equal(t, s3Content, string(content))

	assert.NotNil(t, s3Client.params)
	assert.NotNil(t, bucketName, s3Client.params.Bucket)
	assert.NotNil(t, s3Key, s3Client.params.Key)
}

func Test_fetchS3Content_error(t *testing.T) {
	s3Client := &staticTestS3Client{
		err: errors.New("test-error"),
	}
	record := events.S3EventRecord{
		S3: events.S3Entity{
			Bucket: events.S3Bucket{
				Name: "test-bucket",
			},
			Object: events.S3Object{
				Key: "test-key",
			},
		},
	}
	content, err := fetchS3Content(context.Background(), s3Client, record)
	assert.Error(t, err)
	assert.Nil(t, content)
}
