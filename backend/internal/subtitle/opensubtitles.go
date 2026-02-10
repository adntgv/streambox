package subtitle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/streambox/backend/internal/models"
)

const defaultBaseURL = "https://api.opensubtitles.com/api/v1"

// Client communicates with the OpenSubtitles REST API v1 to search and
// download subtitles.
type Client struct {
	apiKey  string
	http    *http.Client
	baseURL string
}

// NewClient creates an OpenSubtitles client authenticated with the given API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL: defaultBaseURL,
	}
}

// Search finds subtitles for the given IMDb ID and language code (e.g. "en", "ru").
func (c *Client) Search(imdbID string, lang string) ([]models.SubtitleResult, error) {
	reqURL := fmt.Sprintf("%s/subtitles?imdb_id=%s&languages=%s", c.baseURL, imdbID, lang)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build search request: %w", err)
	}
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search subtitles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("opensubtitles api returned status %d", resp.StatusCode)
	}

	var osResp osSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&osResp); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	var results []models.SubtitleResult
	for _, item := range osResp.Data {
		if len(item.Attributes.Files) == 0 {
			continue
		}
		results = append(results, models.SubtitleResult{
			FileID:    item.Attributes.Files[0].FileID,
			Language:  item.Attributes.Language,
			Name:      item.Attributes.Release,
			Downloads: item.Attributes.DownloadCount,
		})
	}

	return results, nil
}

// Download fetches a subtitle file by file ID and returns its contents as
// WebVTT (converted from SRT).
func (c *Client) Download(fileID int) ([]byte, error) {
	// Step 1: Request a download link from the API.
	body, err := json.Marshal(map[string]int{"file_id": fileID})
	if err != nil {
		return nil, fmt.Errorf("marshal download body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/download", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build download request: %w", err)
	}
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request download link: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download api returned status %d", resp.StatusCode)
	}

	var dlResp osDownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&dlResp); err != nil {
		return nil, fmt.Errorf("decode download response: %w", err)
	}

	if dlResp.Link == "" {
		return nil, fmt.Errorf("no download link returned")
	}

	// Step 2: Fetch the actual SRT file from the download link.
	srtResp, err := c.http.Get(dlResp.Link)
	if err != nil {
		return nil, fmt.Errorf("fetch srt file: %w", err)
	}
	defer srtResp.Body.Close()

	if srtResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("srt download returned status %d", srtResp.StatusCode)
	}

	srtData, err := io.ReadAll(srtResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read srt body: %w", err)
	}

	// Step 3: Convert SRT to WebVTT format.
	return srtToVTT(srtData), nil
}

// srtToVTT converts SRT subtitle data to WebVTT format by prepending the
// WEBVTT header and replacing commas with dots in timestamp lines.
func srtToVTT(srt []byte) []byte {
	// Match timestamp lines: "00:01:23,456 --> 00:02:34,789"
	tsRegex := regexp.MustCompile(`(\d{2}:\d{2}:\d{2}),(\d{3})`)
	converted := tsRegex.ReplaceAll(srt, []byte("${1}.${2}"))

	var buf bytes.Buffer
	buf.WriteString("WEBVTT\n\n")
	buf.Write(converted)
	return buf.Bytes()
}

// ----- internal OpenSubtitles response types -----

type osSearchResponse struct {
	Data []osSubtitleItem `json:"data"`
}

type osSubtitleItem struct {
	Attributes osAttributes `json:"attributes"`
}

type osAttributes struct {
	Language      string   `json:"language"`
	Release       string   `json:"release"`
	DownloadCount int      `json:"download_count"`
	Files         []osFile `json:"files"`
}

type osFile struct {
	FileID int `json:"file_id"`
}

type osDownloadResponse struct {
	Link string `json:"link"`
}
