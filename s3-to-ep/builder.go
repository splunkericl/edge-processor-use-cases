package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

const (
	epHostEnvKey                = "EDGE_PROCESSOR_HOST"
	epTLSClientCertEnvKey       = "TLS_CLIENT_CERT"
	epTLSClientPrivateKeyEnvKey = "TLS_CLIENT_KEY"
	epTLSCACertEnvKey           = "TLS_CLIENT_CA_CERT"
	encodingMethodEnvKey        = "ENCODING_METHOD"

	sourcetypeEnvKey = "EVENT_SOURCETYPE"
	indexEnvKey      = "EVENT_INDEX"
	eventIsRawEnvKey = "EVENT_IS_RAW"

	defaultSourcetype = "archived_data"
	defaultIndex      = "main"
	defaultHostName   = "unknownHost"

	formattedEndpointSuffix = "/services/collector"
	rawEndpointSuffix       = "/services/collector/raw"
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

func buildPostBody(isRawEvent bool, record events.S3EventRecord, host, sourcetype, index, s3Content string) ([]byte, error) {
	if isRawEvent {
		return []byte(s3Content), nil
	}

	event := hecEvent{
		Time:       record.EventTime.Unix(),
		Host:       host,
		Source:     record.EventSource,
		Sourcetype: sourcetype,
		Index:      index,
		Event:      s3Content,
	}
	return json.Marshal(event)
}

func buildURL(isRawEvent bool, host, source, sourcetype, index string) (string, error) {
	epHost := os.Getenv(epHostEnvKey)
	if epHost == "" {
		return "", fmt.Errorf("%s has not been provided", epHostEnvKey)
	}

	parsedHostUrl, err := url.Parse(epHost)
	if err != nil {
		return "", err
	}

	if !isRawEvent {
		parsedHostUrl.Path = formattedEndpointSuffix
		return parsedHostUrl.String(), nil
	}

	parsedHostUrl.Path = rawEndpointSuffix
	query := parsedHostUrl.Query()
	query.Set("host", host)
	query.Set("source", source)
	query.Set("sourcetype", sourcetype)
	query.Set("index", index)
	parsedHostUrl.RawQuery = query.Encode()
	return parsedHostUrl.String(), nil
}

func buildHTTPReq(record events.S3EventRecord, s3Content string) (*http.Request, error) {
	host, err := os.Hostname()
	if err != nil {
		host = defaultHostName
	}
	isRawEvent := strings.ToLower(os.Getenv(eventIsRawEnvKey)) == "true"
	sourcetype := getEnvValueOrDefault(sourcetypeEnvKey, defaultSourcetype)
	index := getEnvValueOrDefault(indexEnvKey, defaultIndex)

	epUrl, err := buildURL(isRawEvent, host, record.EventSource, sourcetype, index)
	if err != nil {
		return nil, err
	}

	postBodyBytes, err := buildPostBody(isRawEvent, record, host, sourcetype, index, s3Content)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, epUrl, bytes.NewBuffer(postBodyBytes))
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
