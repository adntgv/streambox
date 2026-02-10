package torrent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/streambox/backend/internal/models"
)

var ytsMirrors = []string{
	"https://yts.mx/api/v2",
	"https://yts.torrentbay.st/api/v2",
}

var ytsBaseURL = ytsMirrors[0]

// Standard torrent trackers for constructing magnet URIs from YTS hashes.
var ytsTrackers = []string{
	"udp://open.demonii.com:1337/announce",
	"udp://tracker.openbittorrent.com:80",
	"udp://tracker.coppersurfer.tk:6969",
	"udp://glotorrents.pw:6969/announce",
	"udp://tracker.opentrackr.org:1337/announce",
	"udp://torrent.gresille.org:80/announce",
	"udp://p4p.arenabg.com:1337",
	"udp://tracker.leechers-paradise.org:6969",
}

// YTS is a torrent provider that uses the YTS.mx API.
type YTS struct {
	client *http.Client
}

func NewYTS() *YTS {
	return &YTS{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (y *YTS) Name() string { return "yts" }

func (y *YTS) Search(title, imdbID string, year string) ([]models.TorrentResult, error) {
	params := url.Values{}
	if imdbID != "" {
		params.Set("query_term", imdbID)
	} else {
		params.Set("query_term", title)
	}

	var resp *http.Response
	var err error
	for _, mirror := range ytsMirrors {
		reqURL := fmt.Sprintf("%s/list_movies.json?%s", mirror, params.Encode())
		resp, err = y.client.Get(reqURL)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("yts api request (all mirrors failed): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yts api returned status %d", resp.StatusCode)
	}

	var ytsResp ytsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&ytsResp); err != nil {
		return nil, fmt.Errorf("decode yts response: %w", err)
	}

	if ytsResp.Status != "ok" || ytsResp.Data.MovieCount == 0 {
		return nil, nil
	}

	var results []models.TorrentResult
	for _, movie := range ytsResp.Data.Movies {
		for _, torr := range movie.Torrents {
			magnet := buildMagnet(torr.Hash, movie.Title)
			results = append(results, models.TorrentResult{
				Provider:  "yts",
				Title:     fmt.Sprintf("%s (%d) [%s] [%s]", movie.Title, movie.Year, torr.Quality, torr.Type),
				MagnetURI: magnet,
				Quality:   strings.ToLower(torr.Quality),
				SizeBytes: torr.SizeBytes,
				SizeHuman: torr.Size,
				Seeds:     torr.Seeds,
				Peers:     torr.Peers,
				Audio:     "English",
				Source:    torr.Type,
			})
		}
	}

	return results, nil
}

func buildMagnet(hash, title string) string {
	magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", hash, url.QueryEscape(title))
	for _, tracker := range ytsTrackers {
		magnet += "&tr=" + url.QueryEscape(tracker)
	}
	return magnet
}

// YTS API response types

type ytsAPIResponse struct {
	Status string  `json:"status"`
	Data   ytsData `json:"data"`
}

type ytsData struct {
	MovieCount int        `json:"movie_count"`
	Movies     []ytsMovie `json:"movies"`
}

type ytsMovie struct {
	Title    string       `json:"title"`
	Year     int          `json:"year"`
	Torrents []ytsTorrent `json:"torrents"`
}

type ytsTorrent struct {
	Hash      string `json:"hash"`
	Quality   string `json:"quality"`
	Type      string `json:"type"`
	Size      string `json:"size"`
	SizeBytes int64  `json:"size_bytes"`
	Seeds     int    `json:"seeds"`
	Peers     int    `json:"peers"`
}
