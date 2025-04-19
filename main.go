package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mallox/qbittorrent-cleaner/qbittorrent"
)

func main() {
	// Get environment variables
	downloadDirsStr := os.Getenv("DOWNLOAD_DIRS")
	downloadDirs := strings.Split(downloadDirsStr, ",")

	serverURL := os.Getenv("SERVER_URL")
	serverUser := os.Getenv("SERVER_USER")
	serverPass := os.Getenv("SERVER_PASS")

	// Create qBittorrent client
	client := qbittorrent.NewClient(serverURL, serverUser, serverPass)

	// Login to qBittorrent
	if err := client.Login(); err != nil {
		fmt.Printf("Failed to login: %v\n", err)
		os.Exit(1)
	}

	// List torrents
	torrents, err := client.ListTorrents()
	if err != nil {
		fmt.Printf("Failed to list torrents: %v\n", err)
		os.Exit(1)
	}

	if len(torrents) == 0 {
		fmt.Println("No torrents found")
		os.Exit(0)
	}

	// Process each torrent
	for _, torrent := range torrents {
		// Skip incomplete torrents unless they're in moving or error state
		if torrent.AmountLeft > 0 && torrent.State != "moving" && torrent.State != "error" {
			fmt.Printf("Skipping because it's not complete: %s\n", torrent.Name)
			continue
		}

		// Get files for this torrent
		files, err := client.TorrentFiles(torrent.Hash)
		if err != nil {
			fmt.Printf("Failed to get files for torrent %s: %v\n", torrent.Name, err)
			continue
		}

		missing := false
		for _, file := range files {
			// Skip files that are not downloaded
			if file.Priority == 0 {
				continue
			}

			// Check if file exists in any download directory
			found := false
			for _, dir := range downloadDirs {
				filePath := filepath.Join(dir, file.Name)
				if _, err := os.Stat(filePath); err == nil {
					found = true
					break
				}
			}

			if !found {
				fmt.Printf("File %s is missing for %s -> REMOVING\n", file.Name, torrent.Name)
				missing = true
				if err := client.RemoveTorrent(torrent.Hash, true); err != nil {
					fmt.Printf("Failed to remove torrent %s: %v\n", torrent.Name, err)
				}
				break
			}
		}

		if !missing {
			fmt.Printf("All files are present for %s\n", torrent.Name)
		}
	}
}
