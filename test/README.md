# Testing Strategy for qbt-clean

This document outlines the testing strategy for the qbt-clean application, focusing on both unit tests and integration tests with an actual qBittorrent instance.

## Testing Goals

1. Verify that the QBittorrentClient correctly interacts with the qBittorrent WebUI API
2. Ensure the application correctly identifies and removes torrents with missing files
3. Validate error handling and edge cases
4. Confirm the application works with an actual qBittorrent instance

## Testing Approach

### Unit Tests

Unit tests will focus on testing individual components of the application in isolation, using mocks for external dependencies.

#### QBittorrentClient Tests

1. **Test NewQBittorrentClient**
   - Verify client is created with correct configuration
   - Check TLS verification is disabled

2. **Test Login**
   - Mock successful login response
   - Mock failed login response (wrong credentials)
   - Mock network error

3. **Test ListTorrents**
   - Mock empty torrents list
   - Mock list with multiple torrents
   - Mock error response
   - Mock malformed JSON response

4. **Test TorrentFiles**
   - Mock empty files list
   - Mock list with multiple files
   - Mock error response
   - Mock malformed JSON response

5. **Test RemoveTorrent**
   - Mock successful removal
   - Mock error response
   - Test with deleteFiles=true and deleteFiles=false

#### Main Function Tests

1. **Test Environment Variable Handling**
   - Test default values
   - Test custom values

2. **Test Torrent Processing Logic**
   - Test skipping incomplete torrents
   - Test handling of "moving" and "error" states
   - Test file existence checking
   - Test torrent removal logic

### Integration Tests

Integration tests will use an actual qBittorrent instance running in a Docker container to verify the application works correctly in a real environment.

#### Setup

1. Use Docker Compose to create a test environment with:
   - qBittorrent container
   - Shared volume for test files
   - qbt-clean application container

2. Configure qBittorrent with:
   - WebUI enabled
   - Test user credentials
   - Test torrents

#### Test Cases

1. **Test Connection to qBittorrent**
   - Verify the application can connect to qBittorrent
   - Test with correct and incorrect credentials

2. **Test Listing Torrents**
   - Add test torrents to qBittorrent
   - Verify the application can list them

3. **Test File Checking**
   - Create test files in the download directory
   - Remove some files
   - Verify the application correctly identifies missing files

4. **Test Torrent Removal**
   - Verify torrents with missing files are removed
   - Verify torrents with all files present are kept

5. **Test Edge Cases**
   - Test with no torrents
   - Test with incomplete torrents
   - Test with torrents in various states (downloading, seeding, paused, etc.)

## Implementation Plan

### 1. Set Up Testing Environment

Create a Docker Compose file that sets up:
- qBittorrent container
- Shared volumes for test files
- Network configuration

### 2. Implement Unit Tests

Create Go test files for each component:
- client_test.go for QBittorrentClient tests
- main_test.go for main function tests

### 3. Implement Integration Tests

Create integration test scripts:
- setup_test_env.sh to set up the test environment
- integration_test.go for Go-based integration tests
- cleanup_test_env.sh to clean up after tests

### 4. Create Test Automation

Create a test.sh script that:
- Runs unit tests
- Sets up the integration test environment
- Runs integration tests
- Cleans up the test environment
- Reports test results

## Test Data

### Sample Torrents

Create sample .torrent files for testing:
- Complete torrent with all files present
- Complete torrent with some files missing
- Incomplete torrent
- Torrent in "moving" state
- Torrent in "error" state

### Sample Files

Create sample files for the torrents:
- Various file sizes
- Various file types
- Files in different directories

## Conclusion

This testing strategy provides a comprehensive approach to testing the qbt-clean application, ensuring it works correctly with an actual qBittorrent instance and handles various edge cases appropriately.