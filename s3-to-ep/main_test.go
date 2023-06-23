package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func Test_handleS3Record_folderEventsAreIgnored(t *testing.T) {
	record := events.S3EventRecord{
		S3: events.S3Entity{
			Bucket: events.S3Bucket{
				Name: "test-bucket",
			},
			Object: events.S3Object{
				Key: "test-key/",
			},
		},
	}
	s3Client := &staticTestS3Client{}
	err := handleS3Record(context.Background(), s3Client, nil, record)
	assert.NoError(t, err)
	assert.Nil(t, s3Client.params)
}

func Test_handleS3Record_hecIngestionSuccess(t *testing.T) {
	const (
		s3Content = "test-content"
		testURL   = "http://localhost/services/collector"
	)
	assert.NoError(t, os.Setenv(epHostEnvKey, testURL))
	t.Cleanup(func() {
		_ = os.Unsetenv(epHostEnvKey)
	})

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
	s3Client := &staticTestS3Client{
		output: &s3.GetObjectOutput{
			Body: io.NopCloser(strings.NewReader(s3Content)),
		},
	}

	tests := []struct {
		name               string
		httpResponseStatus int
		expectedErr        bool
	}{
		{
			name:               "hec event ingestion was successful",
			httpResponseStatus: 200,
		},
		{
			name:               "hec event ingestion ran into server error",
			httpResponseStatus: 500,
			expectedErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(http.MethodPost, testURL, func(_ *http.Request) (*http.Response, error) {
				return httpmock.NewJsonResponse(tt.httpResponseStatus, map[string]interface{}{
					"mockResponse": "",
				})
			})

			err := handleS3Record(context.Background(), s3Client, &http.Client{}, record)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
