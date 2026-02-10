package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/streambox/backend/internal/models"
)

const defaultBaseURL = "https://api.themoviedb.org/3"

// Client communicates with the TMDB v3 API to fetch movie metadata.
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a TMDB client authenticated with the given API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: defaultBaseURL,
	}
}

// Search queries TMDB for movies matching the given query string.
func (c *Client) Search(query string, page int) (*models.MovieSearchResult, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("query", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("language", "ru-RU")
	params.Set("include_adult", "false")

	reqURL := fmt.Sprintf("%s/search/movie?%s", c.baseURL, params.Encode())

	var tmdbResp tmdbSearchResponse
	if err := c.doGet(reqURL, &tmdbResp); err != nil {
		return nil, fmt.Errorf("tmdb search: %w", err)
	}

	result := &models.MovieSearchResult{
		Page:         tmdbResp.Page,
		TotalPages:   tmdbResp.TotalPages,
		TotalResults: tmdbResp.TotalResults,
		Results:      make([]models.Movie, len(tmdbResp.Results)),
	}
	for i, r := range tmdbResp.Results {
		result.Results[i] = r.toMovie()
	}
	return result, nil
}

// GetTrending returns the trending movies for the current week.
func (c *Client) GetTrending() ([]models.Movie, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "ru-RU")

	reqURL := fmt.Sprintf("%s/trending/movie/week?%s", c.baseURL, params.Encode())

	var tmdbResp tmdbSearchResponse
	if err := c.doGet(reqURL, &tmdbResp); err != nil {
		return nil, fmt.Errorf("tmdb trending: %w", err)
	}

	movies := make([]models.Movie, len(tmdbResp.Results))
	for i, r := range tmdbResp.Results {
		movies[i] = r.toMovie()
	}
	return movies, nil
}

// GetPopular returns popular movies from TMDB, paginated.
func (c *Client) GetPopular(page int) (*models.MovieSearchResult, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("page", strconv.Itoa(page))
	params.Set("language", "ru-RU")
	params.Set("include_adult", "false")

	reqURL := fmt.Sprintf("%s/movie/popular?%s", c.baseURL, params.Encode())

	var tmdbResp tmdbSearchResponse
	if err := c.doGet(reqURL, &tmdbResp); err != nil {
		return nil, fmt.Errorf("tmdb popular: %w", err)
	}

	result := &models.MovieSearchResult{
		Page:         tmdbResp.Page,
		TotalPages:   tmdbResp.TotalPages,
		TotalResults: tmdbResp.TotalResults,
		Results:      make([]models.Movie, len(tmdbResp.Results)),
	}
	for i, r := range tmdbResp.Results {
		result.Results[i] = r.toMovie()
	}
	return result, nil
}

// GetDetails returns full movie details including runtime, genres, and IMDb ID.
func (c *Client) GetDetails(id int) (*models.Movie, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "ru-RU")
	params.Set("append_to_response", "external_ids")

	reqURL := fmt.Sprintf("%s/movie/%d?%s", c.baseURL, id, params.Encode())

	var tmdbResp tmdbDetailResponse
	if err := c.doGet(reqURL, &tmdbResp); err != nil {
		return nil, fmt.Errorf("tmdb details for %d: %w", id, err)
	}

	movie := &models.Movie{
		ID:           tmdbResp.ID,
		Title:        tmdbResp.Title,
		Overview:     tmdbResp.Overview,
		PosterPath:   tmdbResp.PosterPath,
		BackdropPath: tmdbResp.BackdropPath,
		ReleaseDate:  tmdbResp.ReleaseDate,
		VoteAverage:  tmdbResp.VoteAverage,
		Runtime:      tmdbResp.Runtime,
		Genres:       make([]models.Genre, len(tmdbResp.Genres)),
	}

	if tmdbResp.ExternalIDs != nil {
		movie.IMDbID = tmdbResp.ExternalIDs.IMDbID
	}

	for i, g := range tmdbResp.Genres {
		movie.Genres[i] = models.Genre{
			ID:   g.ID,
			Name: g.Name,
		}
	}

	return movie, nil
}

// doGet performs an HTTP GET request and JSON-decodes the response body into dest.
func (c *Client) doGet(url string, dest interface{}) error {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tmdb api returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}

// ----- internal TMDB response types -----

type tmdbSearchResponse struct {
	Page         int              `json:"page"`
	TotalPages   int              `json:"total_pages"`
	TotalResults int              `json:"total_results"`
	Results      []tmdbMovieEntry `json:"results"`
}

type tmdbMovieEntry struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	ReleaseDate  string  `json:"release_date"`
	VoteAverage  float64 `json:"vote_average"`
}

func (e *tmdbMovieEntry) toMovie() models.Movie {
	return models.Movie{
		ID:           e.ID,
		Title:        e.Title,
		Overview:     e.Overview,
		PosterPath:   e.PosterPath,
		BackdropPath: e.BackdropPath,
		ReleaseDate:  e.ReleaseDate,
		VoteAverage:  e.VoteAverage,
	}
}

type tmdbDetailResponse struct {
	ID           int              `json:"id"`
	Title        string           `json:"title"`
	Overview     string           `json:"overview"`
	PosterPath   string           `json:"poster_path"`
	BackdropPath string           `json:"backdrop_path"`
	ReleaseDate  string           `json:"release_date"`
	VoteAverage  float64          `json:"vote_average"`
	Runtime      int              `json:"runtime"`
	Genres       []tmdbGenre      `json:"genres"`
	ExternalIDs  *tmdbExternalIDs `json:"external_ids"`
}

type tmdbGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type tmdbExternalIDs struct {
	IMDbID string `json:"imdb_id"`
}
