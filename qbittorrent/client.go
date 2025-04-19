// Package qbittorrent provides a client for the qBittorrent WebUI API
package qbittorrent

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a client for the qBittorrent WebUI API
type Client struct {
	BaseURL  string
	Username string
	Password string
	Client   *http.Client
	Cookies  []*http.Cookie
}

// Torrent represents a torrent in qBittorrent
type Torrent struct {
	Hash       string `json:"hash"`
	Name       string `json:"name"`
	AmountLeft int64  `json:"amount_left"`
	State      string `json:"state"`
}

// TorrentFile represents a file in a torrent
type TorrentFile struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

// NewClient creates a new qBittorrent client
func NewClient(baseURL, username, password string) *Client {
	// Create HTTP client with TLS verification disabled
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	return &Client{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		Client:   client,
	}
}

// Login authenticates with the qBittorrent WebUI
func (c *Client) Login() error {
	data := url.Values{}
	data.Set("username", c.Username)
	data.Set("password", c.Password)

	resp, err := c.Client.PostForm(c.BaseURL+"/api/v2/auth/login", data)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %s", resp.Status)
	}

	c.Cookies = resp.Cookies()

	// If no cookies were returned, we'll use HTTP Basic Authentication as a fallback
	if len(c.Cookies) == 0 {
		fmt.Println("No cookies returned during login, using HTTP Basic Authentication as fallback")
	}

	return nil
}

// ListTorrents returns a list of torrents
func (c *Client) ListTorrents() ([]Torrent, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/v2/torrents/info", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	// Add cookies if available
	for _, cookie := range c.Cookies {
		req.AddCookie(cookie)
	}

	// If no cookies are available, use HTTP Basic Authentication
	if len(c.Cookies) == 0 {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list torrents request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list torrents failed with status: %s, body: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %w", err)
	}

	var torrents []Torrent
	if err := json.Unmarshal(body, &torrents); err != nil {
		return nil, fmt.Errorf("unmarshaling torrents failed: %w", err)
	}

	return torrents, nil
}

// TorrentFiles returns the files for a torrent
func (c *Client) TorrentFiles(hash string) ([]TorrentFile, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/v2/torrents/files?hash="+hash, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	// Add cookies if available
	for _, cookie := range c.Cookies {
		req.AddCookie(cookie)
	}

	// If no cookies are available, use HTTP Basic Authentication
	if len(c.Cookies) == 0 {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("torrent files request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("torrent files failed with status: %s, body: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %w", err)
	}

	var files []TorrentFile
	if err := json.Unmarshal(body, &files); err != nil {
		return nil, fmt.Errorf("unmarshaling files failed: %w", err)
	}

	return files, nil
}

// RemoveTorrent removes a torrent
func (c *Client) RemoveTorrent(hash string, deleteFiles bool) error {
	data := url.Values{}
	data.Set("hashes", hash)
	if deleteFiles {
		data.Set("deleteFiles", "true")
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/v2/torrents/delete", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Add cookies if available
	for _, cookie := range c.Cookies {
		req.AddCookie(cookie)
	}

	// If no cookies are available, use HTTP Basic Authentication
	if len(c.Cookies) == 0 {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("remove torrent request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remove torrent failed with status: %s, body: %s", resp.Status, string(body))
	}

	return nil
}
