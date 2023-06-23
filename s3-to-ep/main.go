package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	httpContentEncodingHeader = "Content-Encoding"
	httpContentTypeHeader     = "Content-Type"
	contentType               = "application/json"
	folderSuffix              = "/"
)

func handleS3Record(ctx context.Context, s3Client S3Client, httpClient *http.Client, record events.S3EventRecord) error {
	// ignore folders
	if strings.HasSuffix(record.S3.Object.Key, folderSuffix) {
		return nil
	}

	s3ContentBytes, err := fetchS3Content(ctx, s3Client, record)
	if err != nil {
		log.Printf("error fetching s3 object: %s", err)
		return err
	}

	// TODO: combine multiple events into a batch to send together
	httpReq, err := buildHTTPReq(record, string(s3ContentBytes))
	if err != nil {
		log.Printf("error building http request: %s", err)
		return err
	}

	res, err := httpClient.Do(httpReq)
	if err != nil {
		log.Printf("error making http call: %s", err)
		return err
	}

	// TODO: we can retry 500s errors with exponential backoffs
	if res.StatusCode >= 400 && res.StatusCode < 600 {
		var responseBody string
		if resBodyBytes, err := io.ReadAll(res.Body); err == nil {
			responseBody = string(resBodyBytes)
		}
		err = fmt.Errorf("http response was not successful. Status code: %d, Response Body: %s", res.StatusCode, responseBody)
		return err
	}
	return nil
}

func S3Handler(ctx context.Context, s3Event events.S3Event) error {
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("failed to load default config: %s", err)
		return err
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	httpClient, err := buildHTTPClient()
	if err != nil {
		log.Printf("error building http client: %s", err)
		return err
	}

	log.Printf("receiving S3 Event records. Count: %d", len(s3Event.Records))
	for _, record := range s3Event.Records {
		if err = handleS3Record(ctx, s3Client, httpClient, record); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	lambda.Start(S3Handler)
}
