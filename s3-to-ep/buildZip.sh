#!/bin/bash

# TODO: handle building in windows properly
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main *.go
zip s3-to-ep.zip main