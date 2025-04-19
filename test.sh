#!/bin/bash
set -e

echo "Building Go application..."
docker build -t qbt-clean-go .

echo "Running Go application..."
docker run --rm \
  -e DOWNLOAD_DIRS=/downloads \
  -e SERVER_URL=https://10.0.0.1:8080 \
  -e SERVER_USER=admin \
  -e SERVER_PASS=adminadmin \
  qbt-clean-go

echo "Test completed successfully!"
