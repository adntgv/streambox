package hdrezka

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
	"github.com/streambox/backend/internal/models"
)

// Client scrapes HDRezka for popular content.
type Client struct {
	mirrors    []string
	httpClient *http.Client
	cache      []models.PopularItem
	cacheTime  time.Time
	mu         sync.RWMutex
}

const cacheDuration = 1 * time.Hour

func NewClient(mirrors ...string) *Client {
	if len(mirrors) == 0 {
		mirrors = []string{"https://hdrezka.ag", "https://rezka.ag"}
	}
	return &Client{
		mirrors: mirrors,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// GetPopular returns the popular items from the HDRezka homepage.
// Results are cached for 1 hour.
func (c *Client) GetPopular() ([]models.PopularItem, error) {
	c.mu.RLock()
	if len(c.cache) > 0 && time.Since(c.cacheTime) < cacheDuration {
		items := c.cache
		c.mu.RUnlock()
		return items, nil
	}
	c.mu.RUnlock()

	var items []models.PopularItem
	var lastErr error

	for _, mirror := range c.mirrors {
		items, lastErr = c.scrapePopular(mirror)
		if lastErr == nil && len(items) > 0 {
			c.mu.Lock()
			c.cache = items
			c.cacheTime = time.Now()
			c.mu.Unlock()

			log.Info().Int("count", len(items)).Str("mirror", mirror).Msg("hdrezka popular loaded")
			return items, nil
		}
		log.Warn().Err(lastErr).Str("mirror", mirror).Msg("hdrezka mirror failed")
	}

	return nil, fmt.Errorf("all hdrezka mirrors failed: %w", lastErr)
}

func (c *Client) scrapePopular(baseURL string) ([]models.PopularItem, error) {
	req, err := http.NewRequest("GET", baseURL+"/", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	var items []models.PopularItem

	doc.Find("div.b-content__inline_item").Each(func(i int, s *goquery.Selection) {
		linkEl := s.Find("div.b-content__inline_item-link a").First()
		title := strings.TrimSpace(linkEl.Text())
		href, _ := linkEl.Attr("href")

		imgEl := s.Find("img").First()
		poster := imgEl.AttrOr("src", "")

		infoEl := s.Find("div.b-content__inline_item-link div").First()
		info := strings.TrimSpace(infoEl.Text())

		if title == "" {
			return
		}

		// Ensure poster URL is absolute
		if poster != "" && !strings.HasPrefix(poster, "http") {
			poster = baseURL + poster
		}
		// Ensure link is absolute
		if href != "" && !strings.HasPrefix(href, "http") {
			href = baseURL + href
		}

		items = append(items, models.PopularItem{
			Title:  title,
			Poster: poster,
			Info:   info,
			URL:    href,
		})
	})

	if len(items) == 0 {
		return nil, fmt.Errorf("no items found on page")
	}

	return items, nil
}
