package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func buildHTTPClientAndAssertNoTLS(t *testing.T) {
	client, err := buildHTTPClient()
	assert.NoError(t, err)
	assert.Nil(t, client.Transport)
}

func Test_buildHTTPClient_noEnvSet_noTLS(t *testing.T) {
	buildHTTPClientAndAssertNoTLS(t)
}

func Test_buildHTTPClient_partialEnvSet_noTLS(t *testing.T) {
	tests := []struct {
		name      string
		preTestFn func()
	}{
		{
			name: "only client cert is set",
			preTestFn: func() {
				err := os.Setenv(epTLSClientCertEnvKey, "test-cert")
				assert.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(epTLSClientCertEnvKey)
				})
			},
		},
		{
			name: "only client key is set",
			preTestFn: func() {
				err := os.Setenv(epTLSClientPrivateKeyEnvKey, "test-key")
				assert.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(epTLSClientPrivateKeyEnvKey)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildHTTPClientAndAssertNoTLS(t)
		})
	}
}

func Test_buildHTTPClient_envSet_hasTLS(t *testing.T) {
	const (
		testCert   = "-----BEGIN CERTIFICATE-----\nMIID0DCCArigAwIBAgIBATANBgkqhkiG9w0BAQUFADB/MQswCQYDVQQGEwJGUjET\nMBEGA1UECAwKU29tZS1TdGF0ZTEOMAwGA1UEBwwFUGFyaXMxDTALBgNVBAoMBERp\nbWkxDTALBgNVBAsMBE5TQlUxEDAOBgNVBAMMB0RpbWkgQ0ExGzAZBgkqhkiG9w0B\nCQEWDGRpbWlAZGltaS5mcjAeFw0xNDAxMjgyMDM2NTVaFw0yNDAxMjYyMDM2NTVa\nMFsxCzAJBgNVBAYTAkZSMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJ\nbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxFDASBgNVBAMMC3d3dy5kaW1pLmZyMIIB\nIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvpnaPKLIKdvx98KW68lz8pGa\nRRcYersNGqPjpifMVjjE8LuCoXgPU0HePnNTUjpShBnynKCvrtWhN+haKbSp+QWX\nSxiTrW99HBfAl1MDQyWcukoEb9Cw6INctVUN4iRvkn9T8E6q174RbcnwA/7yTc7p\n1NCvw+6B/aAN9l1G2pQXgRdYC/+G6o1IZEHtWhqzE97nY5QKNuUVD0V09dc5CDYB\naKjqetwwv6DFk/GRdOSEd/6bW+20z0qSHpa3YNW6qSp+x5pyYmDrzRIR03os6Dau\nZkChSRyc/Whvurx6o85D6qpzywo8xwNaLZHxTQPgcIA5su9ZIytv9LH2E+lSwwID\nAQABo3sweTAJBgNVHRMEAjAAMCwGCWCGSAGG+EIBDQQfFh1PcGVuU1NMIEdlbmVy\nYXRlZCBDZXJ0aWZpY2F0ZTAdBgNVHQ4EFgQU+tugFtyN+cXe1wxUqeA7X+yS3bgw\nHwYDVR0jBBgwFoAUhMwqkbBrGp87HxfvwgPnlGgVR64wDQYJKoZIhvcNAQEFBQAD\nggEBAIEEmqqhEzeXZ4CKhE5UM9vCKzkj5Iv9TFs/a9CcQuepzplt7YVmevBFNOc0\n+1ZyR4tXgi4+5MHGzhYCIVvHo4hKqYm+J+o5mwQInf1qoAHuO7CLD3WNa1sKcVUV\nvepIxc/1aHZrG+dPeEHt0MdFfOw13YdUc2FH6AqEdcEL4aV5PXq2eYR8hR4zKbc1\nfBtuqUsvA8NWSIyzQ16fyGve+ANf6vXvUizyvwDrPRv/kfvLNa3ZPnLMMxU98Mvh\nPXy3PkB8++6U4Y3vdk2Ni2WYYlIls8yqbM4327IKmkDc2TimS8u60CT47mKU7aDY\ncbTV5RDkrlaYwm5yqlTIglvCv7o=\n-----END CERTIFICATE-----"
		testKey    = "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAvpnaPKLIKdvx98KW68lz8pGaRRcYersNGqPjpifMVjjE8LuC\noXgPU0HePnNTUjpShBnynKCvrtWhN+haKbSp+QWXSxiTrW99HBfAl1MDQyWcukoE\nb9Cw6INctVUN4iRvkn9T8E6q174RbcnwA/7yTc7p1NCvw+6B/aAN9l1G2pQXgRdY\nC/+G6o1IZEHtWhqzE97nY5QKNuUVD0V09dc5CDYBaKjqetwwv6DFk/GRdOSEd/6b\nW+20z0qSHpa3YNW6qSp+x5pyYmDrzRIR03os6DauZkChSRyc/Whvurx6o85D6qpz\nywo8xwNaLZHxTQPgcIA5su9ZIytv9LH2E+lSwwIDAQABAoIBAFml8cD9a5pMqlW3\nf9btTQz1sRL4Fvp7CmHSXhvjsjeHwhHckEe0ObkWTRsgkTsm1XLu5W8IITnhn0+1\niNr+78eB+rRGngdAXh8diOdkEy+8/Cee8tFI3jyutKdRlxMbwiKsouVviumoq3fx\nOGQYwQ0Z2l/PvCwy/Y82ffq3ysC5gAJsbBYsCrg14bQo44ulrELe4SDWs5HCjKYb\nEI2b8cOMucqZSOtxg9niLN/je2bo/I2HGSawibgcOdBms8k6TvsSrZMr3kJ5O6J+\n77LGwKH37brVgbVYvbq6nWPL0xLG7dUv+7LWEo5qQaPy6aXb/zbckqLqu6/EjOVe\nydG5JQECgYEA9kKfTZD/WEVAreA0dzfeJRu8vlnwoagL7cJaoDxqXos4mcr5mPDT\nkbWgFkLFFH/AyUnPBlK6BcJp1XK67B13ETUa3i9Q5t1WuZEobiKKBLFm9DDQJt43\nuKZWJxBKFGSvFrYPtGZst719mZVcPct2CzPjEgN3Hlpt6fyw3eOrnoECgYEAxiOu\njwXCOmuGaB7+OW2tR0PGEzbvVlEGdkAJ6TC/HoKM1A8r2u4hLTEJJCrLLTfw++4I\nddHE2dLeR4Q7O58SfLphwgPmLDezN7WRLGr7Vyfuv7VmaHjGuC3Gv9agnhWDlA2Q\ngBG9/R9oVfL0Dc7CgJgLeUtItCYC31bGT3yhV0MCgYEA4k3DG4L+RN4PXDpHvK9I\npA1jXAJHEifeHnaW1d3vWkbSkvJmgVf+9U5VeV+OwRHN1qzPZV4suRI6M/8lK8rA\nGr4UnM4aqK4K/qkY4G05LKrik9Ev2CgqSLQDRA7CJQ+Jn3Nb50qg6hFnFPafN+J7\t\t\n7juWln08wFYV4Atpdd+9XQECgYBxizkZFL+9IqkfOcONvWAzGo+Dq1N0L3J4iTIk\nw56CKWXyj88d4qB4eUU3yJ4uB4S9miaW/eLEwKZIbWpUPFAn0db7i6h3ZmP5ZL8Q\nqS3nQCb9DULmU2/tU641eRUKAmIoka1g9sndKAZuWo+o6fdkIb1RgObk9XNn8R4r\npsv+aQKBgB+CIcExR30vycv5bnZN9EFlIXNKaeMJUrYCXcRQNvrnUIUBvAO8+jAe\nCdLygS5RtgOLZib0IVErqWsP3EI1ACGuLts0vQ9GFLQGaN1SaMS40C9kvns1mlDu\nLhIhYpJ8UsCVt5snWo2N+M+6ANh5tpWdQnEK6zILh4tRbuzaiHgb\n-----END RSA PRIVATE KEY-----"
		testCACert = "-----BEGIN CERTIFICATE-----\nMIID0DCCArigAwIBAgIBATANBgkqhkiG9w0BAQUFADB/MQswCQYDVQQGEwJGUjET\nMBEGA1UECAwKU29tZS1TdGF0ZTEOMAwGA1UEBwwFUGFyaXMxDTALBgNVBAoMBERp\nbWkxDTALBgNVBAsMBE5TQlUxEDAOBgNVBAMMB0RpbWkgQ0ExGzAZBgkqhkiG9w0B\nCQEWDGRpbWlAZGltaS5mcjAeFw0xNDAxMjgyMDM2NTVaFw0yNDAxMjYyMDM2NTVa\nMFsxCzAJBgNVBAYTAkZSMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJ\nbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxFDASBgNVBAMMC3d3dy5kaW1pLmZyMIIB\nIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvpnaPKLIKdvx98KW68lz8pGa\nRRcYersNGqPjpifMVjjE8LuCoXgPU0HePnNTUjpShBnynKCvrtWhN+haKbSp+QWX\nSxiTrW99HBfAl1MDQyWcukoEb9Cw6INctVUN4iRvkn9T8E6q174RbcnwA/7yTc7p\n1NCvw+6B/aAN9l1G2pQXgRdYC/+G6o1IZEHtWhqzE97nY5QKNuUVD0V09dc5CDYB\naKjqetwwv6DFk/GRdOSEd/6bW+20z0qSHpa3YNW6qSp+x5pyYmDrzRIR03os6Dau\nZkChSRyc/Whvurx6o85D6qpzywo8xwNaLZHxTQPgcIA5su9ZIytv9LH2E+lSwwID\nAQABo3sweTAJBgNVHRMEAjAAMCwGCWCGSAGG+EIBDQQfFh1PcGVuU1NMIEdlbmVy\nYXRlZCBDZXJ0aWZpY2F0ZTAdBgNVHQ4EFgQU+tugFtyN+cXe1wxUqeA7X+yS3bgw\nHwYDVR0jBBgwFoAUhMwqkbBrGp87HxfvwgPnlGgVR64wDQYJKoZIhvcNAQEFBQAD\nggEBAIEEmqqhEzeXZ4CKhE5UM9vCKzkj5Iv9TFs/a9CcQuepzplt7YVmevBFNOc0\n+1ZyR4tXgi4+5MHGzhYCIVvHo4hKqYm+J+o5mwQInf1qoAHuO7CLD3WNa1sKcVUV\nvepIxc/1aHZrG+dPeEHt0MdFfOw13YdUc2FH6AqEdcEL4aV5PXq2eYR8hR4zKbc1\nfBtuqUsvA8NWSIyzQ16fyGve+ANf6vXvUizyvwDrPRv/kfvLNa3ZPnLMMxU98Mvh\nPXy3PkB8++6U4Y3vdk2Ni2WYYlIls8yqbM4327IKmkDc2TimS8u60CT47mKU7aDY\ncbTV5RDkrlaYwm5yqlTIglvCv7o=\n-----END CERTIFICATE-----"
	)

	err := os.Setenv(epTLSClientCertEnvKey, testCert)
	assert.NoError(t, err)
	err = os.Setenv(epTLSClientPrivateKeyEnvKey, testKey)
	assert.NoError(t, err)
	err = os.Setenv(epTLSCACertEnvKey, testCACert)
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Unsetenv(epTLSClientCertEnvKey)
		_ = os.Unsetenv(epTLSClientPrivateKeyEnvKey)
		_ = os.Unsetenv(epTLSCACertEnvKey)
	})

	client, err := buildHTTPClient()
	assert.NoError(t, err)
	assert.NotNil(t, client.Transport)

	castedTransport, isHTTPTransport := client.Transport.(*http.Transport)
	assert.True(t, isHTTPTransport)
	assert.NotNil(t, castedTransport.TLSClientConfig)
	assert.NotEmpty(t, castedTransport.TLSClientConfig.Certificates)
	assert.NotNil(t, castedTransport.TLSClientConfig.RootCAs)
}

func Test_buildHTTPReq_noCustomizedEnv_success(t *testing.T) {
	const (
		epHost       = "http://localhost"
		source       = "test-source"
		eventContent = "s3-content"
	)
	assert.NoError(t, os.Setenv(epHostEnvKey, epHost))
	t.Cleanup(func() {
		_ = os.Unsetenv(epHostEnvKey)
	})

	curTime := time.Time{}
	record := events.S3EventRecord{
		EventSource: source,
		EventTime:   curTime,
	}

	req, err := buildHTTPReq(record, eventContent)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	assert.Equal(t, contentType, req.Header.Get(httpContentTypeHeader))
	assert.Equal(t, http.MethodPost, req.Method)
	assert.NotNil(t, req.URL)
	assert.Equal(t, epHost+formattedEndpointSuffix, req.URL.String())

	body, err := io.ReadAll(req.Body)
	assert.NoError(t, err)

	var event hecEvent
	err = json.Unmarshal(body, &event)
	assert.NoError(t, err)

	expectedHost, err := os.Hostname()
	assert.NoError(t, err)

	assert.Equal(t, hecEvent{
		Time:       curTime.Unix(),
		Host:       expectedHost,
		Source:     source,
		Sourcetype: defaultSourcetype,
		Index:      defaultIndex,
		Event:      eventContent,
	}, event)
}

func Test_buildHTTPReq_allCustomizedEnv_success(t *testing.T) {
	const (
		testURL          = "http://localhost"
		source           = "test-source"
		customSourcetype = "test-sourcetype"
		customIndex      = "test-index"
		eventContent     = "s3-content"
	)
	assert.NoError(t, os.Setenv(epHostEnvKey, testURL))
	assert.NoError(t, os.Setenv(sourcetypeEnvKey, customSourcetype))
	assert.NoError(t, os.Setenv(indexEnvKey, customIndex))
	assert.NoError(t, os.Setenv(encodingMethodEnvKey, gzipEncoding))
	t.Cleanup(func() {
		_ = os.Unsetenv(epHostEnvKey)
		_ = os.Unsetenv(sourcetypeEnvKey)
		_ = os.Unsetenv(indexEnvKey)
		_ = os.Unsetenv(encodingMethodEnvKey)
	})

	curTime := time.Time{}
	record := events.S3EventRecord{
		EventSource: source,
		EventTime:   curTime,
	}

	req, err := buildHTTPReq(record, eventContent)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	assert.Equal(t, contentType, req.Header.Get(httpContentTypeHeader))
	assert.Equal(t, gzipEncoding, req.Header.Get(httpContentEncodingHeader))
	assert.Equal(t, http.MethodPost, req.Method)
	assert.NotNil(t, req.URL)
	assert.Equal(t, testURL+formattedEndpointSuffix, req.URL.String())

	body, err := io.ReadAll(req.Body)
	assert.NoError(t, err)

	var event hecEvent
	err = json.Unmarshal(body, &event)
	assert.NoError(t, err)

	expectedHost, err := os.Hostname()
	assert.NoError(t, err)

	assert.Equal(t, hecEvent{
		Time:       curTime.Unix(),
		Host:       expectedHost,
		Source:     source,
		Sourcetype: customSourcetype,
		Index:      customIndex,
		Event:      eventContent,
	}, event)
}

func Test_buildHTTPReq_rawEvent_success(t *testing.T) {
	const (
		epHost       = "http://localhost"
		source       = "test-source"
		eventContent = "s3-content"
	)
	assert.NoError(t, os.Setenv(epHostEnvKey, epHost))
	assert.NoError(t, os.Setenv(eventIsRawEnvKey, "true"))
	t.Cleanup(func() {
		_ = os.Unsetenv(epHostEnvKey)
		_ = os.Unsetenv(eventIsRawEnvKey)
	})

	curTime := time.Time{}
	record := events.S3EventRecord{
		EventSource: source,
		EventTime:   curTime,
	}

	req, err := buildHTTPReq(record, eventContent)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	assert.Equal(t, contentType, req.Header.Get(httpContentTypeHeader))
	assert.Equal(t, http.MethodPost, req.Method)

	expectedHost, err := os.Hostname()
	assert.NoError(t, err)

	// assert URL
	assert.NotNil(t, req.URL)
	assert.Equal(t, "localhost", req.URL.Host)
	assert.Equal(t, rawEndpointSuffix, req.URL.Path)
	queries := req.URL.Query()
	assert.Equal(t, defaultSourcetype, queries.Get("sourcetype"))
	assert.Equal(t, expectedHost, queries.Get("host"))
	assert.Equal(t, record.EventSource, queries.Get("source"))
	assert.Equal(t, defaultIndex, queries.Get("index"))

	body, err := io.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.Equal(t, eventContent, string(body))
}

func Test_buildHTTPReq_unexpectedEncodingType_error(t *testing.T) {
	assert.NoError(t, os.Setenv(epHostEnvKey, "http://www.splunk.com"))
	assert.NoError(t, os.Setenv(encodingMethodEnvKey, "compress"))
	t.Cleanup(func() {
		_ = os.Unsetenv(epHostEnvKey)
		_ = os.Unsetenv(encodingMethodEnvKey)
	})

	req, err := buildHTTPReq(events.S3EventRecord{}, "s3-content")
	assert.Error(t, err)
	assert.Equal(t, "compress is not supported. Only GZIP is supported", err.Error())
	assert.Nil(t, req)
}
