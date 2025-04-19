#!/bin/bash
set -e

# Function to print section headers
print_header() {
  echo "============================================"
  echo "$1"
  echo "============================================"
}

# Function to run unit tests
run_unit_tests() {
  print_header "Running unit tests"
  cd ..
  go test ./qbittorrent -v
  cd test
}

# Function to run integration tests
run_integration_tests() {
  print_header "Setting up integration test environment"

  # Create test data directories
  mkdir -p test-data/config
  mkdir -p test-data/downloads

  # Create sample files for testing
  echo "Creating sample files for testing..."
  mkdir -p test-data/downloads/complete_torrent
  echo "This is a sample file" > test-data/downloads/complete_torrent/sample_file.txt
  dd if=/dev/urandom of=test-data/downloads/complete_torrent/large_file.bin bs=1M count=10 2>/dev/null

  # Create incomplete torrent directory
  mkdir -p test-data/downloads/incomplete_torrent
  echo "This is an incomplete file" > test-data/downloads/incomplete_torrent/incomplete_file.txt

  # Download a sample torrent file for testing
  echo "Downloading sample torrent files..."
  mkdir -p test-data/torrents
  curl -L "https://releases.ubuntu.com/22.04/ubuntu-22.04.3-desktop-amd64.iso.torrent" -o test-data/torrents/ubuntu.torrent

  # Start the test environment
  print_header "Starting qBittorrent container"
  docker-compose up -d qbittorrent

  # Wait for qBittorrent to start
  echo "Waiting for qBittorrent to start..."
  sleep 10

  # Add torrents to qBittorrent
  echo "Adding torrents to qBittorrent..."
  # This would typically use the qBittorrent API to add torrents
  # For now, we'll just print a message
  echo "Note: You need to manually add torrents to qBittorrent for testing"
  echo "You can use the WebUI at http://localhost:8080 (admin/adminadmin)"

  # Run integration tests
  print_header "Running integration tests"
  cd ..
  INTEGRATION_TEST=true DOWNLOAD_DIR=./test/test-data/downloads SERVER_URL=http://localhost:8080 go test ./test -v
  cd test

  # Stop the test environment
  print_header "Stopping test environment"
  docker-compose down

  # Clean up test data
  echo "Cleaning up test data..."
  rm -rf test-data
}

# Main script
print_header "Starting test suite"

# Run unit tests
run_unit_tests

# Check if integration tests should be run
if [[ $1 == "--integration" || $1 == "-i" ]]; then
  run_integration_tests
else
  echo "Skipping integration tests. Use --integration or -i flag to run them."
fi

print_header "Test suite completed"
