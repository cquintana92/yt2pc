package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	playlistCache = make(map[string]*PlaylistCacheItem)
	cacheMutex    sync.RWMutex
	cacheExpiry   time.Duration // Cache expiration duration
)

type Server struct {
	apiUrl        string
	youtubeApiKey string
	filterPattern string
}

// RSSFeed represents the structure of the RSS feed
type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Link        string    `xml:"link"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Enclosure   struct {
		URL    string `xml:"url,attr"`
		Length string `xml:"length,attr"`
		Type   string `xml:"type,attr"`
	} `xml:"enclosure"`
}

// Cache for playlist data
type PlaylistCacheItem struct {
	PlaylistItems []*youtube.PlaylistItem
	FetchedAt     time.Time
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handleRequest(w, r)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/health" {
		s.healthCheck(w)
		return
	}

	if len(path) < 2 {
		http.NotFound(w, r)
		return
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && strings.HasSuffix(parts[0], ".xml") {
		slug := strings.TrimSuffix(parts[0], ".xml")
		log.Printf("Received RSS feed request for slug: %s", slug)
		s.serveRSSFeed(w, r, slug)
	} else if len(parts) == 2 {
		slug, videoID := parts[0], parts[1]
		log.Printf("Received audio request for slug: %s, videoID: %s", slug, videoID)
		serveAudio(w, r, videoID)
	} else {
		http.NotFound(w, r)
	}
}

func (s *Server) healthCheck(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) serveRSSFeed(w http.ResponseWriter, r *http.Request, slug string) {

	playlistItems, err := s.getPlaylistItemsCached(slug, s.youtubeApiKey, r)
	if err != nil {
		http.Error(w, "Error fetching playlist items", http.StatusInternalServerError)
		log.Printf("Error fetching playlist items for slug %s: %v", slug, err)
		return
	}

	filteredItems := s.filterVideos(playlistItems)
	rssFeed := s.generateRSSFeed(filteredItems, slug)

	w.Header().Set("Content-Type", "application/rss+xml")
	xmlData, err := xml.MarshalIndent(rssFeed, "", "  ")
	if err != nil {
		http.Error(w, "Error generating RSS feed", http.StatusInternalServerError)
		log.Printf("Error generating RSS feed for slug %s: %v", slug, err)
		return
	}
	w.Write([]byte(xml.Header))
	w.Write(xmlData)
	log.Printf("Served RSS feed for slug: %s", slug)
}

// Get playlist items with cache handling
func (s *Server) getPlaylistItemsCached(slug string, apiKey string, r *http.Request) ([]*youtube.PlaylistItem, error) {
	cacheMutex.RLock()
	cacheItem, cached := playlistCache[slug]
	cacheMutex.RUnlock()

	if cached && time.Since(cacheItem.FetchedAt) < cacheExpiry {
		log.Printf("Using cached playlist items for slug: %s", slug)
		return cacheItem.PlaylistItems, nil
	}

	log.Printf("Fetching playlist items from YouTube API for slug: %s", slug)
	ytService, err := youtube.NewService(r.Context(), option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Error creating YouTube service: %v", err)
		return nil, err
	}

	playlistItems, err := fetchPlaylistItems(ytService, slug)
	if err != nil {
		log.Printf("Error fetching playlist items for slug %s: %v", slug, err)
		return nil, err
	}

	// Update the cache
	cacheMutex.Lock()
	playlistCache[slug] = &PlaylistCacheItem{
		PlaylistItems: playlistItems,
		FetchedAt:     time.Now(),
	}
	cacheMutex.Unlock()
	log.Printf("Updated cache for slug: %s", slug)

	return playlistItems, nil
}

func fetchPlaylistItems(ytService *youtube.Service, playlistID string) ([]*youtube.PlaylistItem, error) {
	var allItems []*youtube.PlaylistItem
	nextPageToken := ""

	for {
		call := ytService.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(playlistID).
			MaxResults(50).
			PageToken(nextPageToken)

		response, err := call.Do()
		if err != nil {
			log.Printf("Error fetching playlist items from YouTube API: %v", err)
			return nil, err
		}

		allItems = append(allItems, response.Items...)
		if response.NextPageToken == "" {
			break
		}
		nextPageToken = response.NextPageToken
	}

	log.Printf("Fetched %d playlist items from YouTube API for playlistID: %s", len(allItems), playlistID)
	return allItems, nil
}

func (s *Server) filterVideos(items []*youtube.PlaylistItem) []*youtube.PlaylistItem {
	if s.filterPattern == "" {
		return items
	}

	var filtered []*youtube.PlaylistItem
	regex := regexp.MustCompile(s.filterPattern)

	for _, item := range items {
		title := item.Snippet.Title
		if regex.MatchString(title) {
			filtered = append(filtered, item)
		}
	}
	log.Printf("Filtered %d videos out of %d using pattern: %s", len(filtered), len(items), s.filterPattern)

	// Revert the list so new episodes show first
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	return filtered
}

func (s *Server) generateRSSFeed(items []*youtube.PlaylistItem, slug string) RSSFeed {
	var rssItems []RSSItem
	for _, item := range items {
		videoID := item.Snippet.ResourceId.VideoId
		rssItem := RSSItem{
			Title:       item.Snippet.Title,
			Description: item.Snippet.Description,
			Link:        fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
			GUID:        videoID,
			Enclosure: struct {
				URL    string `xml:"url,attr"`
				Length string `xml:"length,attr"`
				Type   string `xml:"type,attr"`
			}{
				URL:    fmt.Sprintf("%s/%s/%s", s.apiUrl, slug, videoID),
				Length: "0",
				Type:   "audio/mpeg",
			},
		}
		rssItems = append(rssItems, rssItem)
	}

	feed := RSSFeed{
		Version: "2.0",
		Channel: RSSChannel{
			Title:       fmt.Sprintf("YouTube Playlist %s", slug),
			Description: "Generated podcast feed from YouTube playlist",
			Link:        fmt.Sprintf("%s/%s.xml", s.apiUrl, slug),
			Items:       rssItems,
		},
	}
	log.Printf("Generated RSS feed with %d items for slug: %s", len(rssItems), slug)
	return feed
}

func serveAudio(w http.ResponseWriter, r *http.Request, videoID string) {
	audioFilePath := filepath.Join(audioDir, fmt.Sprintf("%s.mp3", videoID))

	if _, err := os.Stat(audioFilePath); os.IsNotExist(err) {
		log.Printf("Audio file not found in cache, downloading videoID: %s", videoID)
		err := downloadAudio(videoID, audioFilePath)
		if err != nil {
			http.Error(w, "Error downloading audio", http.StatusInternalServerError)
			log.Printf("Error downloading audio for videoID %s: %v", videoID, err)
			return
		}
	} else {
		log.Printf("Serving cached audio file for videoID: %s", videoID)
	}

	// Open the audio file
	audioFile, err := os.Open(audioFilePath)
	if err != nil {
		http.Error(w, "Error opening audio file", http.StatusInternalServerError)
		log.Printf("Error opening audio file for videoID %s: %v", videoID, err)
		return
	}
	defer audioFile.Close()

	// Get file info
	stat, err := audioFile.Stat()
	if err != nil {
		http.Error(w, "Error getting file info", http.StatusInternalServerError)
		log.Printf("Error getting file info for videoID %s: %v", videoID, err)
		return
	}

	// Serve the file with support for Range requests
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), audioFile)
}

func downloadAudio(videoID string, outputPath string) error {
	cmd := exec.Command("yt-dlp", "-f", "bestaudio", "--extract-audio", "--audio-format", "mp3", "-o", outputPath, fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID))
	log.Printf("Running yt-dlp command for videoID: %s", videoID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("yt-dlp error for videoID %s: %v\nOutput: %s", videoID, err, string(output))
		return err
	}
	log.Printf("Successfully downloaded audio for videoID: %s", videoID)
	return nil
}
