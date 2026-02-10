package models

type Movie struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Overview    string   `json:"overview"`
	PosterPath  string   `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	ReleaseDate string   `json:"release_date"`
	VoteAverage float64  `json:"vote_average"`
	Runtime     int      `json:"runtime"`
	IMDbID      string   `json:"imdb_id,omitempty"`
	Genres      []Genre  `json:"genres,omitempty"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MovieSearchResult struct {
	Page         int     `json:"page"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
	Results      []Movie `json:"results"`
}

type TorrentResult struct {
	Provider  string `json:"provider"`
	Title     string `json:"title"`
	MagnetURI string `json:"magnet_uri"`
	Quality   string `json:"quality"`
	SizeBytes int64  `json:"size_bytes"`
	SizeHuman string `json:"size_human"`
	Seeds     int    `json:"seeds"`
	Peers     int    `json:"peers"`
	Audio     string `json:"audio"`
	Source    string `json:"source"`
	TopicID   string `json:"topic_id,omitempty"`
}

type AudioTrack struct {
	Index    int    `json:"index"`
	Language string `json:"language"`
	Title    string `json:"title"`
}

type StreamSession struct {
	ID             string       `json:"session_id"`
	TMDbID         int          `json:"tmdb_id"`
	Title          string       `json:"title"`
	MagnetURI      string       `json:"magnet_uri"`
	InfoHash       string       `json:"info_hash"`
	FilePath       string       `json:"file_path,omitempty"`
	FileSize       int64        `json:"file_size"`
	ContentType    string       `json:"content_type"`
	NeedsTranscode bool         `json:"needs_transcode"`
	Status         string       `json:"status"`
	Duration       float64      `json:"duration"`
	AudioTracks    []AudioTrack `json:"audio_tracks,omitempty"`
}

type StreamStatus struct {
	Status          string       `json:"status"`
	DownloadedBytes int64        `json:"downloaded_bytes"`
	TotalBytes      int64        `json:"total_bytes"`
	DownloadSpeed   int64        `json:"download_speed"`
	PeersConnected  int          `json:"peers_connected"`
	BufferedPercent float64      `json:"buffered_percent"`
	Duration        float64      `json:"duration"`
	AudioTracks     []AudioTrack `json:"audio_tracks,omitempty"`
}

type WatchHistory struct {
	ID         int     `json:"id"`
	TMDbID     int     `json:"tmdb_id"`
	Title      string  `json:"title"`
	PosterPath string  `json:"poster_path"`
	Year       int     `json:"year"`
	Duration   int     `json:"duration"`
	Progress   float64 `json:"progress"`
	Completed  bool    `json:"completed"`
	Quality    string  `json:"quality"`
	MagnetURI  string  `json:"magnet_uri"`
	WatchedAt  string  `json:"watched_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type SubtitleResult struct {
	FileID   int    `json:"file_id"`
	Language string `json:"language"`
	Name     string `json:"name"`
	Downloads int   `json:"downloads"`
}

// ----- TV Series types -----

type TVShow struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Overview        string    `json:"overview"`
	PosterPath      string    `json:"poster_path"`
	BackdropPath    string    `json:"backdrop_path"`
	FirstAirDate    string    `json:"first_air_date"`
	VoteAverage     float64   `json:"vote_average"`
	NumberOfSeasons int       `json:"number_of_seasons,omitempty"`
	NumberOfEpisodes int      `json:"number_of_episodes,omitempty"`
	IMDbID          string    `json:"imdb_id,omitempty"`
	Genres          []Genre   `json:"genres,omitempty"`
	Seasons         []Season  `json:"seasons,omitempty"`
}

type Season struct {
	ID            int       `json:"id"`
	SeasonNumber  int       `json:"season_number"`
	Name          string    `json:"name"`
	Overview      string    `json:"overview"`
	PosterPath    string    `json:"poster_path"`
	AirDate       string    `json:"air_date"`
	EpisodeCount  int       `json:"episode_count"`
	Episodes      []Episode `json:"episodes,omitempty"`
}

type Episode struct {
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

type TVShowSearchResult struct {
	Page         int      `json:"page"`
	TotalPages   int      `json:"total_pages"`
	TotalResults int      `json:"total_results"`
	Results      []TVShow `json:"results"`
}

// MediaItem is a unified type for mixed movie/TV content.
type MediaItem struct {
	ID           int     `json:"id"`
	MediaType    string  `json:"media_type"`
	Title        string  `json:"title"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	Date         string  `json:"date"`
	VoteAverage  float64 `json:"vote_average"`
}

type MediaSearchResult struct {
	Page         int         `json:"page"`
	TotalPages   int         `json:"total_pages"`
	TotalResults int         `json:"total_results"`
	Results      []MediaItem `json:"results"`
}

// TorrentFile represents a single file inside a multi-file torrent.
type TorrentFile struct {
	Index     int    `json:"index"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	SizeHuman string `json:"size_human"`
}
