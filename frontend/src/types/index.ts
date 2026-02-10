export interface Genre {
  id: number
  name: string
}

export interface Movie {
  id: number
  title: string
  overview: string
  poster_path: string
  backdrop_path: string
  release_date: string
  vote_average: number
  runtime: number
  imdb_id: string
  genres: Genre[]
}

export interface MovieSearchResult {
  page: number
  total_pages: number
  total_results: number
  results: Movie[]
}

export interface TorrentResult {
  provider: string
  title: string
  magnet_uri: string
  quality: string
  size_bytes: number
  size_human: string
  seeds: number
  peers: number
  audio: string
  source: string
  topic_id: number
}

export interface AudioTrack {
  index: number
  language: string
  title: string
}

export interface StreamSession {
  session_id: string
  tmdb_id: number
  title: string
  magnet_uri: string
  info_hash: string
  file_path: string
  file_size: number
  content_type: string
  needs_transcode: boolean
  status: string
  duration: number
  audio_tracks?: AudioTrack[]
}

export interface StreamStatus {
  status: string
  downloaded_bytes: number
  total_bytes: number
  download_speed: number
  peers_connected: number
  buffered_percent: number
  duration: number
  audio_tracks?: AudioTrack[]
}

export interface WatchHistory {
  id: number
  tmdb_id: number
  title: string
  poster_path: string
  year: number
  duration: number
  progress: number
  completed: boolean
  quality: string
  magnet_uri: string
  watched_at: string
  updated_at: string
}

export interface SubtitleResult {
  file_id: number
  language: string
  name: string
  downloads: number
}

// --- TV Series ---

export interface TVShow {
  id: number
  name: string
  overview: string
  poster_path: string
  backdrop_path: string
  first_air_date: string
  vote_average: number
  number_of_seasons?: number
  number_of_episodes?: number
  imdb_id?: string
  genres?: Genre[]
  seasons?: Season[]
}

export interface Season {
  id: number
  season_number: number
  name: string
  overview: string
  poster_path: string
  air_date: string
  episode_count: number
  episodes?: Episode[]
}

export interface Episode {
  id: number
  episode_number: number
  season_number: number
  name: string
  overview: string
  still_path: string
  air_date: string
  vote_average: number
  runtime: number
}

export interface TVShowSearchResult {
  page: number
  total_pages: number
  total_results: number
  results: TVShow[]
}

// --- Unified Media ---

export interface MediaItem {
  id: number
  media_type: 'movie' | 'tv'
  title: string
  overview: string
  poster_path: string
  backdrop_path: string
  date: string
  vote_average: number
}

export interface MediaSearchResult {
  page: number
  total_pages: number
  total_results: number
  results: MediaItem[]
}

export interface TorrentFile {
  index: number
  path: string
  size: number
  size_human: string
}
