package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
)

const (
	epHECEndpointEnvKey         = "EDGE_PROCESSOR_HEC_ENDPOINT"
	epTLSClientCertEnvKey       = "TLS_CLIENT_CERT"
	epTLSClientPrivateKeyEnvKey = "TLS_CLIENT_KEY"
	epTLSCACertEnvKey           = "TLS_CLIENT_CA_CERT"
	encodingMethodEnvKey        = "ENCODING_METHOD"

	sourcetypeEnvKey = "EVENT_SOURCETYPE"
	indexEnvKey      = "EVENT_INDEX"

	defaultSourcetype = "archived_data"
	defaultIndex      = "main"
	defaultHostName   = "unknownHost"
)

type hecEvent struct {
	Time       int64  `json:"Time"`
	Host       string `json:"Host"`
	Source     string `json:"Source"`
	Sourcetype string `json:"Sourcetype"`
	Index      string `json:"Index"`
	Event      string `json:"Event"`
}

func buildHTTPClient() (*http.Client, error) {
	client := &http.Client{}
	clientCert := os.Getenv(epTLSClientCertEnvKey)
	clientKey := os.Getenv(epTLSClientPrivateKeyEnvKey)

	if clientCert != "" && clientKey != "" {
		cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
		if err != nil {
			return nil, err
		}

		certPool, err := x509.SystemCertPool()
		if err != nil {
			certPool = x509.NewCertPool()
		}
		caCert := os.Getenv(epTLSCACertEnvKey)
		if caCert != "" {
			certPool.AppendCertsFromPEM([]byte(caCert))
		}

		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      certPool,
			},
		}
	}
	return client, nil
}

func buildHTTPReq(record events.S3EventRecord, s3Content string) (*http.Request, error) {
	epURL := os.Getenv(epHECEndpointEnvKey)
	if epURL == "" {
		return nil, fmt.Errorf("%s has not been provided", epHECEndpointEnvKey)
	}

	host, err := os.Hostname()
	if err != nil {
		host = defaultHostName
	}

	sourcetype := getEnvValueOrDefault(sourcetypeEnvKey, defaultSourcetype)
	index := getEnvValueOrDefault(indexEnvKey, defaultIndex)

	event := hecEvent{
		Time:       record.EventTime.Unix(),
		Host:       host,
		Source:     record.EventSource,
		Sourcetype: sourcetype,
		Index:      index,
		Event:      s3Content,
	}
	serializedEvent, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, epURL, bytes.NewBuffer(serializedEvent))
	if err != nil {
		return nil, err
	}

	encodingMethod := os.Getenv(encodingMethodEnvKey)
	if encodingMethod != "" {
		if encodingMethod != gzipEncoding {
			return nil, fmt.Errorf("%s is not supported. Only GZIP is supported", encodingMethod)
		}
		req.Header.Set(httpContentEncodingHeader, encodingMethod)
	}
	req.Header.Set(httpContentTypeHeader, contentType)

	return req, nil
}
