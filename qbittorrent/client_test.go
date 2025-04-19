package qbittorrent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewClient tests the creation of a new Client
func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8080", "admin", "adminadmin")

	if client.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected BaseURL to be 'http://localhost:8080', got '%s'", client.BaseURL)
	}

	if client.Username != "admin" {
		t.Errorf("Expected Username to be 'admin', got '%s'", client.Username)
	}

	if client.Password != "adminadmin" {
		t.Errorf("Expected Password to be 'adminadmin', got '%s'", client.Password)
	}

	if client.Client == nil {
		t.Error("Expected Client to be initialized, got nil")
	}
}

// TestLogin tests the Login method
func TestLogin(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			err := r.ParseForm()
			if err != nil {
				t.Fatalf("Failed to parse form: %v", err)
			}

			username := r.Form.Get("username")
			password := r.Form.Get("password")

			if username == "admin" && password == "adminadmin" {
				// Set a cookie for successful login
				http.SetCookie(w, &http.Cookie{
					Name:  "SID",
					Value: "test-session-id",
				})
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		} else {
			t.Errorf("Unexpected request to %s", r.URL.Path)
		}
	}))
	defer server.Close()

	// Test successful login
	client := NewClient(server.URL, "admin", "adminadmin")
	err := client.Login()
	if err != nil {
		t.Errorf("Expected successful login, got error: %v", err)
	}

	if len(client.Cookies) == 0 {
		t.Error("Expected cookies to be set after login")
	}

	// Test failed login
	client = NewClient(server.URL, "wrong", "credentials")
	err = client.Login()
	if err == nil {
		t.Error("Expected login to fail with wrong credentials")
	}
}

// TestListTorrents tests the ListTorrents method
func TestListTorrents(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			// Handle login request
			http.SetCookie(w, &http.Cookie{
				Name:  "SID",
				Value: "test-session-id",
			})
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == "/api/v2/torrents/info" {
			// Check for cookie
			cookie, err := r.Cookie("SID")
			if err != nil || cookie.Value != "test-session-id" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Return sample torrents
			torrents := []Torrent{
				{
					Hash:       "abcdef123456",
					Name:       "Test Torrent 1",
					AmountLeft: 0,
					State:      "completed",
				},
				{
					Hash:       "123456abcdef",
					Name:       "Test Torrent 2",
					AmountLeft: 1024,
					State:      "downloading",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(torrents)
		} else {
			t.Errorf("Unexpected request to %s", r.URL.Path)
		}
	}))
	defer server.Close()

	// Create client and login
	client := NewClient(server.URL, "admin", "adminadmin")
	err := client.Login()
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Test listing torrents
	torrents, err := client.ListTorrents()
	if err != nil {
		t.Errorf("Failed to list torrents: %v", err)
	}

	if len(torrents) != 2 {
		t.Errorf("Expected 2 torrents, got %d", len(torrents))
	}

	if torrents[0].Name != "Test Torrent 1" {
		t.Errorf("Expected first torrent name to be 'Test Torrent 1', got '%s'", torrents[0].Name)
	}

	if torrents[1].AmountLeft != 1024 {
		t.Errorf("Expected second torrent amount_left to be 1024, got %d", torrents[1].AmountLeft)
	}
}

// TestTorrentFiles tests the TorrentFiles method
func TestTorrentFiles(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			// Handle login request
			http.SetCookie(w, &http.Cookie{
				Name:  "SID",
				Value: "test-session-id",
			})
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == "/api/v2/torrents/files" {
			// Check for cookie
			cookie, err := r.Cookie("SID")
			if err != nil || cookie.Value != "test-session-id" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Check for hash parameter
			hash := r.URL.Query().Get("hash")
			if hash != "abcdef123456" {
				t.Errorf("Expected hash parameter to be 'abcdef123456', got '%s'", hash)
			}

			// Return sample files
			files := []TorrentFile{
				{
					Name:     "file1.txt",
					Priority: 1,
				},
				{
					Name:     "file2.txt",
					Priority: 0,
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(files)
		} else {
			t.Errorf("Unexpected request to %s", r.URL.Path)
		}
	}))
	defer server.Close()

	// Create client and login
	client := NewClient(server.URL, "admin", "adminadmin")
	err := client.Login()
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Test getting torrent files
	files, err := client.TorrentFiles("abcdef123456")
	if err != nil {
		t.Errorf("Failed to get torrent files: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	if files[0].Name != "file1.txt" {
		t.Errorf("Expected first file name to be 'file1.txt', got '%s'", files[0].Name)
	}

	if files[1].Priority != 0 {
		t.Errorf("Expected second file priority to be 0, got %d", files[1].Priority)
	}
}

// TestRemoveTorrent tests the RemoveTorrent method
func TestRemoveTorrent(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			// Handle login request
			http.SetCookie(w, &http.Cookie{
				Name:  "SID",
				Value: "test-session-id",
			})
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == "/api/v2/torrents/delete" {
			// Check for cookie
			cookie, err := r.Cookie("SID")
			if err != nil || cookie.Value != "test-session-id" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Check method
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			// Parse form
			err = r.ParseForm()
			if err != nil {
				t.Fatalf("Failed to parse form: %v", err)
			}

			// Check parameters
			hashes := r.Form.Get("hashes")
			if hashes != "abcdef123456" {
				t.Errorf("Expected hashes parameter to be 'abcdef123456', got '%s'", hashes)
			}

			deleteFiles := r.Form.Get("deleteFiles")
			if deleteFiles != "true" {
				t.Errorf("Expected deleteFiles parameter to be 'true', got '%s'", deleteFiles)
			}

			w.WriteHeader(http.StatusOK)
		} else {
			t.Errorf("Unexpected request to %s", r.URL.Path)
		}
	}))
	defer server.Close()

	// Create client and login
	client := NewClient(server.URL, "admin", "adminadmin")
	err := client.Login()
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Test removing torrent
	err = client.RemoveTorrent("abcdef123456", true)
	if err != nil {
		t.Errorf("Failed to remove torrent: %v", err)
	}
}
