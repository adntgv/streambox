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
