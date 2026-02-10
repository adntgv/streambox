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

// ----- TV Series methods -----

// SearchTV queries TMDB for TV shows matching the given query string.
func (c *Client) SearchTV(query string, page int) (*models.TVShowSearchResult, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("query", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("language", "ru-RU")

	reqURL := fmt.Sprintf("%s/search/tv?%s", c.baseURL, params.Encode())

	var tmdbResp tmdbTVSearchResponse
	if err := c.doGet(reqURL, &tmdbResp); err != nil {
		return nil, fmt.Errorf("tmdb search tv: %w", err)
	}

	result := &models.TVShowSearchResult{
		Page:         tmdbResp.Page,
		TotalPages:   tmdbResp.TotalPages,
		TotalResults: tmdbResp.TotalResults,
		Results:      make([]models.TVShow, len(tmdbResp.Results)),
	}
	for i, r := range tmdbResp.Results {
		result.Results[i] = r.toTVShow()
	}
	return result, nil
}

// GetTrendingTV returns the trending TV shows for the current week.
func (c *Client) GetTrendingTV() ([]models.TVShow, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "ru-RU")

	reqURL := fmt.Sprintf("%s/trending/tv/week?%s", c.baseURL, params.Encode())

	var tmdbResp tmdbTVSearchResponse
	if err := c.doGet(reqURL, &tmdbResp); err != nil {
		return nil, fmt.Errorf("tmdb trending tv: %w", err)
	}

	shows := make([]models.TVShow, len(tmdbResp.Results))
	for i, r := range tmdbResp.Results {
		shows[i] = r.toTVShow()
	}
	return shows, nil
}

// GetPopularTV returns popular TV shows from TMDB, paginated.
func (c *Client) GetPopularTV(page int) (*models.TVShowSearchResult, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("page", strconv.Itoa(page))
	params.Set("language", "ru-RU")

	reqURL := fmt.Sprintf("%s/tv/popular?%s", c.baseURL, params.Encode())

	var tmdbResp tmdbTVSearchResponse
	if err := c.doGet(reqURL, &tmdbResp); err != nil {
		return nil, fmt.Errorf("tmdb popular tv: %w", err)
	}

	result := &models.TVShowSearchResult{
		Page:         tmdbResp.Page,
		TotalPages:   tmdbResp.TotalPages,
		TotalResults: tmdbResp.TotalResults,
		Results:      make([]models.TVShow, len(tmdbResp.Results)),
	}
	for i, r := range tmdbResp.Results {
		result.Results[i] = r.toTVShow()
	}
	return result, nil
}

// GetTVDetails returns full TV show details including seasons and IMDb ID.
func (c *Client) GetTVDetails(id int) (*models.TVShow, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "ru-RU")
	params.Set("append_to_response", "external_ids")

	reqURL := fmt.Sprintf("%s/tv/%d?%s", c.baseURL, id, params.Encode())

	var tmdbResp tmdbTVDetailResponse
	if err := c.doGet(reqURL, &tmdbResp); err != nil {
		return nil, fmt.Errorf("tmdb tv details for %d: %w", id, err)
	}

	show := &models.TVShow{
		ID:               tmdbResp.ID,
		Name:             tmdbResp.Name,
		Overview:         tmdbResp.Overview,
		PosterPath:       tmdbResp.PosterPath,
		BackdropPath:     tmdbResp.BackdropPath,
		FirstAirDate:     tmdbResp.FirstAirDate,
		VoteAverage:      tmdbResp.VoteAverage,
		NumberOfSeasons:  tmdbResp.NumberOfSeasons,
		NumberOfEpisodes: tmdbResp.NumberOfEpisodes,
		Genres:           make([]models.Genre, len(tmdbResp.Genres)),
		Seasons:          make([]models.Season, len(tmdbResp.Seasons)),
	}

	if tmdbResp.ExternalIDs != nil {
		show.IMDbID = tmdbResp.ExternalIDs.IMDbID
	}

	for i, g := range tmdbResp.Genres {
		show.Genres[i] = models.Genre{ID: g.ID, Name: g.Name}
	}

	for i, s := range tmdbResp.Seasons {
		show.Seasons[i] = models.Season{
			ID:           s.ID,
			SeasonNumber: s.SeasonNumber,
			Name:         s.Name,
			Overview:     s.Overview,
			PosterPath:   s.PosterPath,
			AirDate:      s.AirDate,
			EpisodeCount: s.EpisodeCount,
		}
	}

	return show, nil
}

// GetSeasonDetails returns full season details including all episodes.
func (c *Client) GetSeasonDetails(tvID, seasonNumber int) (*models.Season, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "ru-RU")

	reqURL := fmt.Sprintf("%s/tv/%d/season/%d?%s", c.baseURL, tvID, seasonNumber, params.Encode())

	var tmdbResp tmdbSeasonDetailResponse
	if err := c.doGet(reqURL, &tmdbResp); err != nil {
		return nil, fmt.Errorf("tmdb season %d for tv %d: %w", seasonNumber, tvID, err)
	}

	season := &models.Season{
		ID:           tmdbResp.ID,
		SeasonNumber: tmdbResp.SeasonNumber,
		Name:         tmdbResp.Name,
		Overview:     tmdbResp.Overview,
		PosterPath:   tmdbResp.PosterPath,
		AirDate:      tmdbResp.AirDate,
		EpisodeCount: len(tmdbResp.Episodes),
		Episodes:     make([]models.Episode, len(tmdbResp.Episodes)),
	}

	for i, e := range tmdbResp.Episodes {
		season.Episodes[i] = models.Episode{
			ID:            e.ID,
			EpisodeNumber: e.EpisodeNumber,
			SeasonNumber:  e.SeasonNumber,
			Name:          e.Name,
			Overview:      e.Overview,
			StillPath:     e.StillPath,
			AirDate:       e.AirDate,
			VoteAverage:   e.VoteAverage,
			Runtime:       e.Runtime,
		}
	}

	return season, nil
}

// SearchMulti queries TMDB for both movies and TV shows, filtering out person results.
func (c *Client) SearchMulti(query string, page int) (*models.MediaSearchResult, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("query", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("language", "ru-RU")
	params.Set("include_adult", "false")

	reqURL := fmt.Sprintf("%s/search/multi?%s", c.baseURL, params.Encode())

	var tmdbResp tmdbMultiSearchResponse
	if err := c.doGet(reqURL, &tmdbResp); err != nil {
		return nil, fmt.Errorf("tmdb search multi: %w", err)
	}

	var items []models.MediaItem
	for _, r := range tmdbResp.Results {
		if r.MediaType == "movie" || r.MediaType == "tv" {
			items = append(items, r.toMediaItem())
		}
	}

	return &models.MediaSearchResult{
		Page:         tmdbResp.Page,
		TotalPages:   tmdbResp.TotalPages,
		TotalResults: tmdbResp.TotalResults,
		Results:      items,
	}, nil
}

// GetTrendingAll returns trending movies and TV shows for the current week.
func (c *Client) GetTrendingAll() ([]models.MediaItem, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "ru-RU")

	reqURL := fmt.Sprintf("%s/trending/all/week?%s", c.baseURL, params.Encode())

	var tmdbResp tmdbMultiSearchResponse
	if err := c.doGet(reqURL, &tmdbResp); err != nil {
		return nil, fmt.Errorf("tmdb trending all: %w", err)
	}

	var items []models.MediaItem
	for _, r := range tmdbResp.Results {
		if r.MediaType == "movie" || r.MediaType == "tv" {
			items = append(items, r.toMediaItem())
		}
	}
	return items, nil
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

// ----- TV series internal types -----

type tmdbTVEntry struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	FirstAirDate string  `json:"first_air_date"`
	VoteAverage  float64 `json:"vote_average"`
}

func (e *tmdbTVEntry) toTVShow() models.TVShow {
	return models.TVShow{
		ID:           e.ID,
		Name:         e.Name,
		Overview:     e.Overview,
		PosterPath:   e.PosterPath,
		BackdropPath: e.BackdropPath,
		FirstAirDate: e.FirstAirDate,
		VoteAverage:  e.VoteAverage,
	}
}

type tmdbTVSearchResponse struct {
	Page         int           `json:"page"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
	Results      []tmdbTVEntry `json:"results"`
}

type tmdbTVDetailResponse struct {
	ID               int              `json:"id"`
	Name             string           `json:"name"`
	Overview         string           `json:"overview"`
	PosterPath       string           `json:"poster_path"`
	BackdropPath     string           `json:"backdrop_path"`
	FirstAirDate     string           `json:"first_air_date"`
	VoteAverage      float64          `json:"vote_average"`
	NumberOfSeasons  int              `json:"number_of_seasons"`
	NumberOfEpisodes int              `json:"number_of_episodes"`
	Genres           []tmdbGenre      `json:"genres"`
	Seasons          []tmdbSeason     `json:"seasons"`
	ExternalIDs      *tmdbExternalIDs `json:"external_ids"`
}

type tmdbSeason struct {
	ID           int    `json:"id"`
	SeasonNumber int    `json:"season_number"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	AirDate      string `json:"air_date"`
	EpisodeCount int    `json:"episode_count"`
}

type tmdbSeasonDetailResponse struct {
	ID           int            `json:"id"`
	SeasonNumber int            `json:"season_number"`
	Name         string         `json:"name"`
	Overview     string         `json:"overview"`
	PosterPath   string         `json:"poster_path"`
	AirDate      string         `json:"air_date"`
	Episodes     []tmdbEpisode  `json:"episodes"`
}

type tmdbEpisode struct {
	ID            int     `json:"id"`
	EpisodeNumber int     `json:"episode_number"`
	SeasonNumber  int     `json:"season_number"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	StillPath     string  `json:"still_path"`
	AirDate       string  `json:"air_date"`
	VoteAverage   float64 `json:"vote_average"`
	Runtime       int     `json:"runtime"`
}

type tmdbMultiEntry struct {
	ID           int     `json:"id"`
	MediaType    string  `json:"media_type"`
	Title        string  `json:"title"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	ReleaseDate  string  `json:"release_date"`
	FirstAirDate string  `json:"first_air_date"`
	VoteAverage  float64 `json:"vote_average"`
}

func (e *tmdbMultiEntry) toMediaItem() models.MediaItem {
	title := e.Title
	date := e.ReleaseDate
	if e.MediaType == "tv" {
		title = e.Name
		date = e.FirstAirDate
	}
	return models.MediaItem{
		ID:           e.ID,
		MediaType:    e.MediaType,
		Title:        title,
		Overview:     e.Overview,
		PosterPath:   e.PosterPath,
		BackdropPath: e.BackdropPath,
		Date:         date,
		VoteAverage:  e.VoteAverage,
	}
}

type tmdbMultiSearchResponse struct {
	Page         int              `json:"page"`
	TotalPages   int              `json:"total_pages"`
	TotalResults int              `json:"total_results"`
	Results      []tmdbMultiEntry `json:"results"`
}
