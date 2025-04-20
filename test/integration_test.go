package test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mallox/qbittorrent-cleaner/qbittorrent"
)

// uploadTorrent uploads a torrent file to qBittorrent
func uploadTorrent(t *testing.T, serverURL, username, password, torrentPath string) error {
	// Create a buffer to store the multipart form data
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add the torrent file to the form
	f, err := os.Open(torrentPath)
	if err != nil {
		return fmt.Errorf("failed to open torrent file: %v", err)
	}
	defer f.Close()

	fw, err := w.CreateFormFile("torrents", filepath.Base(torrentPath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err = io.Copy(fw, f); err != nil {
		return fmt.Errorf("failed to copy file data: %v", err)
	}

	// Close the writer to finalize the form
	w.Close()

	// Create the HTTP request
	req, err := http.NewRequest("POST", serverURL+"/api/v2/torrents/add", &b)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the content type
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Create a simple cookie jar for authentication
	jar := &CustomCookieJar{
		Cookies: []*http.Cookie{},
	}

	// Login to get the authentication cookie
	loginURL := serverURL + "/api/v2/auth/login"
	loginData := fmt.Sprintf("username=%s&password=%s", username, password)
	loginReq, err := http.NewRequest("POST", loginURL, strings.NewReader(loginData))
	if err != nil {
		return fmt.Errorf("failed to create login request: %v", err)
	}

	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the login request
	client := &http.Client{}
	loginResp, err := client.Do(loginReq)
	if err != nil {
		return fmt.Errorf("failed to send login request: %v", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %s", loginResp.Status)
	}

	// Extract the cookies
	for _, cookie := range loginResp.Cookies() {
		jar.Cookies = append(jar.Cookies, cookie)
	}

	// Add the cookies to the upload request
	for _, cookie := range jar.Cookies {
		req.AddCookie(cookie)
	}

	// Send the upload request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send upload request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status: %s, body: %s", resp.Status, string(body))
	}

	return nil
}

// CustomCookieJar is a simple cookie jar implementation
type CustomCookieJar struct {
	Cookies []*http.Cookie
}

// TestIntegration tests the qBittorrent client against an actual qBittorrent instance
func TestIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	// Get qBittorrent server URL from environment or use default
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		// Try to use the container name as the hostname when running in Docker
		if _, err := os.Stat("/.dockerenv"); err == nil {
			serverURL = "http://qbittorrent:8080"
		} else {
			serverURL = "http://localhost:8080"
		}
	}

	// Print the server URL for debugging
	fmt.Printf("Using qBittorrent server URL: %s\n", serverURL)

	// Get qBittorrent credentials from environment or use default
	serverUser := os.Getenv("SERVER_USER")
	if serverUser == "" {
		serverUser = "admin"
	}

	// Use the temporary password from the logs
	serverPass := os.Getenv("SERVER_PASS")
	if serverPass == "" {
		serverPass = "zprhcA5fh" // Temporary password from the logs
	}

	// Get download directory from environment or use default
	downloadDir := os.Getenv("DOWNLOAD_DIR")
	if downloadDir == "" {
		downloadDir = "./test-data/downloads"
	}

	// Create qBittorrent client
	client := qbittorrent.NewClient(serverURL, serverUser, serverPass)

	// Test login
	t.Log("Testing login...")
	t.Logf("Using server URL: %s", serverURL)
	t.Logf("Using credentials: %s / %s", serverUser, serverPass)
	err := client.Login()
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	t.Log("Login successful")
	t.Logf("Cookies: %v", client.Cookies)

	// Test listing torrents
	t.Log("Testing listing torrents...")
	torrents, err := client.ListTorrents()
	if err != nil {
		// Try to get more information about the error
		req, _ := http.NewRequest("GET", serverURL+"/api/v2/torrents/info", nil)
		for _, cookie := range client.Cookies {
			req.AddCookie(cookie)
		}
		resp, reqErr := http.DefaultClient.Do(req)
		if reqErr != nil {
			t.Logf("Manual request failed: %v", reqErr)
		} else {
			defer resp.Body.Close()
			t.Logf("Manual request status: %s", resp.Status)
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Manual request body: %s", string(body))
		}

		t.Fatalf("Failed to list torrents: %v", err)
	}
	t.Logf("Found %d torrents", len(torrents))

	// If there are no torrents, we can't test further
	if len(torrents) == 0 {
		t.Log("No torrents found. Please add torrents to qBittorrent for testing.")
		return
	}

	// Test getting files for each torrent
	for _, torrent := range torrents {
		t.Logf("Testing getting files for torrent %s...", torrent.Name)
		files, err := client.TorrentFiles(torrent.Hash)
		if err != nil {
			t.Errorf("Failed to get files for torrent %s: %v", torrent.Name, err)
			continue
		}
		t.Logf("Found %d files for torrent %s", len(files), torrent.Name)

		// Check if files exist in download directory
		for _, file := range files {
			if file.Priority == 0 {
				// Skip files that are not downloaded
				continue
			}

			filePath := filepath.Join(downloadDir, file.Name)
			_, err := os.Stat(filePath)
			if err != nil {
				t.Errorf("File %s does not exist in download directory: %v", file.Name, err)
			} else {
				t.Logf("File %s exists in download directory", file.Name)
			}
		}
	}

	// Test removing a torrent (only if we have a test torrent)
	testTorrentHash := os.Getenv("TEST_TORRENT_HASH")
	if testTorrentHash != "" {
		t.Logf("Testing removing torrent %s...", testTorrentHash)

		// Verify the torrent exists before removal
		torrents, err := client.ListTorrents()
		if err != nil {
			t.Fatalf("Failed to list torrents: %v", err)
		}

		torrentExists := false
		for _, torrent := range torrents {
			if torrent.Hash == testTorrentHash {
				torrentExists = true
				break
			}
		}

		if !torrentExists {
			t.Fatalf("Test torrent %s does not exist before removal", testTorrentHash)
		}

		// Remove the torrent
		err = client.RemoveTorrent(testTorrentHash, false) // Don't delete files
		if err != nil {
			t.Errorf("Failed to remove torrent %s: %v", testTorrentHash, err)
		} else {
			t.Logf("Successfully removed torrent %s", testTorrentHash)
		}

		// Verify the torrent was removed
		time.Sleep(1 * time.Second) // Wait a bit for the removal to take effect

		torrents, err = client.ListTorrents()
		if err != nil {
			t.Fatalf("Failed to list torrents after removal: %v", err)
		}

		for _, torrent := range torrents {
			if torrent.Hash == testTorrentHash {
				t.Errorf("Torrent %s still exists after removal", testTorrentHash)
				break
			}
		}
	} else {
		t.Log("Skipping torrent removal test. Set TEST_TORRENT_HASH to test.")
	}
}

// TestMain is the entry point for integration tests
func TestMain(m *testing.M) {
	// Check if we're running integration tests
	if os.Getenv("INTEGRATION_TEST") != "true" {
		// Just run the tests (which will be skipped)
		os.Exit(m.Run())
	}

	// Wait for qBittorrent to start
	fmt.Println("Waiting for qBittorrent to start...")
	time.Sleep(10 * time.Second)

	// Run the tests
	exitCode := m.Run()

	// Exit with the test exit code
	os.Exit(exitCode)
}
