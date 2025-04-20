# qbt-clean

A utility to automatically remove torrents from qBittorrent when their files are missing.

## Description

This application connects to a qBittorrent WebUI, checks all completed torrents, and removes any torrents whose files are missing from the specified download directories.

## Features

- Connects to qBittorrent WebUI using API
- Checks all completed torrents
- Skips incomplete torrents (unless they're in "moving" or "error" state)
- Checks if files exist in specified download directories
- Removes torrents with missing files
- Logs status of each torrent

## Docker Image Optimization

The Docker image has been optimized to reduce its footprint while maintaining full functionality. The following optimizations were made:

### Original Dockerfile
- Used golang:1.24-alpine as the builder image
- Used alpine:3.18 as the final image
- Total image size: 14.4MB

### Optimized Dockerfile
- Used golang:1.24-alpine as the builder image
- Used scratch as the final image (the smallest possible base image)
- Added build flags for size optimization:
  - `-ldflags="-s -w"` to strip debug information
  - `-trimpath` to remove file paths from the binary
- Specified GOARCH=amd64 for a more targeted build
- Added CA certificates for HTTPS requests
- Changed CMD to ENTRYPOINT for a more direct execution
- Total image size: 5.12MB (64% reduction)

### Further Optimized Dockerfile with UPX
- Added UPX compression to the binary
  - Used `--best --lzma` options for maximum compression
- Enhanced build flags with static linking:
  - `-ldflags="-s -w -extldflags '-static'"`
- Improved documentation with detailed comments
- Total image size: 1.83MB (87% reduction from original, 64% reduction from optimized)

## Environment Variables

- `DOWNLOAD_DIRS`: Comma-separated list of directories to check for downloaded files (default: /downloads)
- `SERVER_URL`: URL of the qBittorrent server (default: https://10.0.0.1:8080)
- `SERVER_USER`: Username for the qBittorrent server (default: admin)
- `SERVER_PASS`: Password for the qBittorrent server (default: adminadmin)

## Usage

```bash
docker run -v /path/to/downloads:/downloads -e SERVER_URL=https://your-qbittorrent-server:8080 -e SERVER_USER=your-username -e SERVER_PASS=your-password qbt-clean
```

## Notes

- The application disables TLS certificate verification to allow connecting to qBittorrent instances with self-signed certificates.
- The application is designed to be run periodically (e.g., via cron) to clean up torrents with missing files.
- The application will terminate after checking all torrents.

## Testing

The application includes both unit tests and integration tests.

### Unit Tests

Unit tests use mock HTTP responses to test the qBittorrent client without requiring an actual qBittorrent instance.

To run unit tests:

```bash
cd test
./run_tests.sh
# Answer 'n' when asked about integration tests
```

Or directly:

```bash
go test ./qbittorrent -v
```

### Integration Tests

Integration tests interact with an actual qBittorrent instance running in a Docker container.

To run integration tests:

```bash
cd test
./run_tests.sh --integration
```

Or directly:

```bash
# Start the qBittorrent container
cd test
docker-compose up -d qbittorrent

# Wait for qBittorrent to start
sleep 10

# Run integration tests
cd ..
INTEGRATION_TEST=true DOWNLOAD_DIR=./test/test-data/downloads SERVER_URL=http://localhost:8080 go test ./test -v

# Stop the qBittorrent container
cd test
docker-compose down
```

### Test Environment

The integration tests use a Docker Compose setup with:
- A qBittorrent container
- Shared volumes for test data
- Network configuration

The test environment is automatically set up and torn down by the test script.

## License

MIT
