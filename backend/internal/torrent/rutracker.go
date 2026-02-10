package torrent

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
	"github.com/streambox/backend/internal/models"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Rutracker category IDs for movies (HD video, UHD, Russian films, animation).
const rutrackerMovieCategories = "313,312,2198,2199,1950,2540"

// Rutracker category IDs for TV series (foreign HD/SD, Russian series).
const rutrackerTVCategories = "189,2366,2100,911,1531,2370,2102,2104,2109"

// Movie forum category names to filter results (in case extra categories slip in).
var movieForumKeywords = []string{
	"кино", "фильм", "video", "uhd", "remux", "3d",
}

// TV forum category keywords.
var tvForumKeywords = []string{
	"сериал", "serial", "series", "season", "сезон", "tv",
}

// Rutracker is a torrent search provider that scrapes rutracker.org.
type Rutracker struct {
	mirror   string
	username string
	password string
	client   *http.Client
	loggedIn bool
}

func NewRutracker(mirror, username, password string) *Rutracker {
	jar, _ := cookiejar.New(nil)
	return &Rutracker{
		mirror:   mirror,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}

func (r *Rutracker) Name() string { return "rutracker" }

// login authenticates with Rutracker and stores the session cookie.
func (r *Rutracker) login() error {
	loginURL := fmt.Sprintf("https://%s/forum/login.php", r.mirror)

	data := url.Values{
		"login_username": {r.username},
		"login_password": {r.password},
		"login":          {"Вход"},
	}

	resp, err := r.client.PostForm(loginURL, data)
	if err != nil {
		return fmt.Errorf("rutracker login request: %w", err)
	}
	defer resp.Body.Close()

	// Check for bb_session cookie in response headers directly
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "bb_session" {
			r.loggedIn = true
			log.Info().Msg("rutracker login successful")
			return nil
		}
	}

	// Also check the cookie jar (redirect may have consumed the cookies)
	u, _ := url.Parse(fmt.Sprintf("https://%s/forum/", r.mirror))
	for _, cookie := range r.client.Jar.Cookies(u) {
		if cookie.Name == "bb_session" {
			r.loggedIn = true
			log.Info().Msg("rutracker login successful")
			return nil
		}
	}

	return fmt.Errorf("rutracker login failed: bb_session cookie not found")
}

func (r *Rutracker) ensureLoggedIn() error {
	if !r.loggedIn {
		return r.login()
	}
	return nil
}

// Search searches Rutracker for movie torrents matching the given title.
func (r *Rutracker) Search(title, imdbID string, year string) ([]models.TorrentResult, error) {
	query := title
	if year != "" {
		query += " " + year
	}
	return r.doSearch(query, rutrackerMovieCategories, movieForumKeywords)
}

// SearchTV searches Rutracker for TV series torrents.
func (r *Rutracker) SearchTV(title string, seasonNum int, year string) ([]models.TorrentResult, error) {
	query := title
	if seasonNum > 0 {
		query += fmt.Sprintf(" сезон %d", seasonNum)
	}
	return r.doSearch(query, rutrackerTVCategories, tvForumKeywords)
}

// doSearch is the shared search logic for both movies and TV.
func (r *Rutracker) doSearch(query, categories string, forumKeywords []string) ([]models.TorrentResult, error) {
	if err := r.ensureLoggedIn(); err != nil {
		return nil, err
	}

	searchURL := fmt.Sprintf("https://%s/forum/tracker.php?nm=%s&c=%s",
		r.mirror, url.QueryEscape(query), categories)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build search request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		r.loggedIn = false
		if loginErr := r.login(); loginErr != nil {
			return nil, loginErr
		}
		resp, err = r.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("rutracker search: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rutracker returned status %d", resp.StatusCode)
	}

	// Decode cp1251 → UTF-8
	utf8Reader := transform.NewReader(resp.Body, charmap.Windows1251.NewDecoder())
	doc, err := goquery.NewDocumentFromReader(utf8Reader)
	if err != nil {
		return nil, fmt.Errorf("parse search results: %w", err)
	}

	results := r.parseSearchResults(doc, forumKeywords)

	// Fetch magnet links for top results (limit to avoid too many requests)
	limit := 10
	if len(results) < limit {
		limit = len(results)
	}
	for i := 0; i < limit; i++ {
		if results[i].TopicID != "" {
			magnet, err := r.getMagnet(results[i].TopicID)
			if err != nil {
				log.Warn().Err(err).Str("topic", results[i].TopicID).Msg("failed to get magnet")
				continue
			}
			results[i].MagnetURI = magnet
		}
	}

	// Filter out results without magnets
	var filtered []models.TorrentResult
	for _, res := range results {
		if res.MagnetURI != "" {
			filtered = append(filtered, res)
		}
	}

	return filtered, nil
}

// parseSearchResults extracts torrent results from the Rutracker HTML table.
func (r *Rutracker) parseSearchResults(doc *goquery.Document, forumKeywords []string) []models.TorrentResult {
	var results []models.TorrentResult

	doc.Find("tr.hl-tr").Each(func(i int, s *goquery.Selection) {
		// Topic title and link
		titleLink := s.Find("a.tLink")
		topicTitle := strings.TrimSpace(titleLink.Text())
		if topicTitle == "" {
			return
		}

		// Extract topic ID from data attribute
		topicID, _ := s.Attr("data-topic_id")
		if topicID == "" {
			topicID = extractTopicID(titleLink.AttrOr("href", ""))
		}

		// Forum category — filter by keywords
		forumName := strings.ToLower(s.Find("td.f-name-col a").Text())
		matched := len(forumKeywords) == 0 // if no keywords, accept all
		for _, kw := range forumKeywords {
			if strings.Contains(forumName, kw) {
				matched = true
				break
			}
		}
		if !matched {
			return
		}

		// Seeds — <b class="seedmed">
		seedsText := strings.TrimSpace(s.Find("b.seedmed").Text())
		seeds, _ := strconv.Atoi(seedsText)

		// Leeches
		leechText := strings.TrimSpace(s.Find("td.leechmed").Text())
		peers, _ := strconv.Atoi(leechText)

		// Size — from data-ts_text attribute (bytes)
		sizeTD := s.Find("td.tor-size")
		sizeAttr, _ := sizeTD.Attr("data-ts_text")
		sizeBytes, _ := strconv.ParseInt(sizeAttr, 10, 64)
		sizeHuman := formatSize(sizeBytes)

		// Parse title for quality, audio info, source
		quality := extractQuality(topicTitle)
		audio := extractAudio(topicTitle)
		source := extractSource(topicTitle)

		results = append(results, models.TorrentResult{
			Provider:  "rutracker",
			Title:     topicTitle,
			Quality:   quality,
			SizeBytes: sizeBytes,
			SizeHuman: sizeHuman,
			Seeds:     seeds,
			Peers:     peers,
			Audio:     audio,
			Source:    source,
			TopicID:   topicID,
		})
	})

	return results
}

// getMagnet fetches the magnet link from a topic page.
func (r *Rutracker) getMagnet(topicID string) (string, error) {
	topicURL := fmt.Sprintf("https://%s/forum/viewtopic.php?t=%s", r.mirror, topicID)

	resp, err := r.client.Get(topicURL)
	if err != nil {
		return "", fmt.Errorf("fetch topic: %w", err)
	}
	defer resp.Body.Close()

	// Read body and decode cp1251
	utf8Reader := transform.NewReader(resp.Body, charmap.Windows1251.NewDecoder())
	body, err := io.ReadAll(utf8Reader)
	if err != nil {
		return "", fmt.Errorf("read topic page: %w", err)
	}

	// Extract magnet link with regex (faster than full HTML parse)
	magnetRe := regexp.MustCompile(`magnet:\?xt=urn:btih:[a-fA-F0-9]+[^"'\s]*`)
	match := magnetRe.Find(body)
	if match == nil {
		return "", fmt.Errorf("no magnet link found on topic %s", topicID)
	}

	return string(match), nil
}

func extractTopicID(href string) string {
	re := regexp.MustCompile(`t=(\d+)`)
	matches := re.FindStringSubmatch(href)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

var qualityRe = regexp.MustCompile(`(?i)\b(2160p|1080p|720p|480p|4K|UHD)\b`)

func extractQuality(title string) string {
	match := qualityRe.FindString(title)
	if match == "" {
		return "unknown"
	}
	return strings.ToLower(match)
}

var audioPatterns = []struct {
	pattern *regexp.Regexp
	label   string
}{
	{regexp.MustCompile(`(?i)\bDub\b`), "Дубляж"},
	{regexp.MustCompile(`(?i)Дубляж`), "Дубляж"},
	{regexp.MustCompile(`(?i)Дублированный`), "Дубляж"},
	{regexp.MustCompile(`(?i)Лицензия`), "Лицензия"},
	{regexp.MustCompile(`(?i)\bDVO\b`), "DVO"},
	{regexp.MustCompile(`(?i)\bAVO\b`), "AVO"},
	{regexp.MustCompile(`(?i)\bUkr\b`), "Ukr"},
	{regexp.MustCompile(`(?i)Original\s*\(Eng\)`), "Original Eng"},
	{regexp.MustCompile(`(?i)Проф\.\s*(?:много|одно)голос`), "Профессиональный перевод"},
	{regexp.MustCompile(`(?i)iTunes`), "iTunes"},
}

func extractAudio(title string) string {
	var found []string
	seen := make(map[string]bool)
	for _, ap := range audioPatterns {
		if ap.pattern.MatchString(title) && !seen[ap.label] {
			found = append(found, ap.label)
			seen[ap.label] = true
		}
	}
	if len(found) == 0 {
		return ""
	}
	return strings.Join(found, ", ")
}

var sourceRe = regexp.MustCompile(`(?i)\b(BDRip|BDRemux|WEB-DL|WEB-DLRip|WEBRip|HDRip|DVDRip|HDTVRip|Blu-ray|BluRay|UHD BDRip)\b`)

func extractSource(title string) string {
	return sourceRe.FindString(title)
}

func formatSize(bytes int64) string {
	if bytes <= 0 {
		return ""
	}
	const (
		gb = 1024 * 1024 * 1024
		mb = 1024 * 1024
	)
	if bytes >= gb {
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	}
	return fmt.Sprintf("%.0f MB", float64(bytes)/float64(mb))
}
