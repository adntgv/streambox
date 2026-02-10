import type {
  Movie,
  MovieSearchResult,
  TorrentResult,
  StreamSession,
  StreamStatus,
  SubtitleResult,
  WatchHistory,
  TVShow,
  TVShowSearchResult,
  Season,
  MediaItem,
  MediaSearchResult,
  TorrentFile,
} from '../types'

const BASE = '/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
    },
    ...options,
  })
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText)
    throw new Error(`API error ${res.status}: ${text}`)
  }
  return res.json()
}

// --- Movies ---

export async function searchMovies(query: string, page = 1): Promise<MovieSearchResult> {
  return request<MovieSearchResult>(`/movies/search?query=${encodeURIComponent(query)}&page=${page}`)
}

export async function getTrending(): Promise<Movie[]> {
  return request<Movie[]>('/movies/trending')
}

export async function getPopular(page = 1): Promise<MovieSearchResult> {
  return request<MovieSearchResult>(`/movies/popular?page=${page}`)
}

export async function getMovieDetails(id: number): Promise<Movie> {
  return request<Movie>(`/movies/${id}`)
}

// --- TV Shows ---

export async function searchTV(query: string, page = 1): Promise<TVShowSearchResult> {
  return request<TVShowSearchResult>(`/tv/search?q=${encodeURIComponent(query)}&page=${page}`)
}

export async function getTrendingTV(): Promise<TVShow[]> {
  return request<TVShow[]>('/tv/trending')
}

export async function getPopularTV(page = 1): Promise<TVShowSearchResult> {
  return request<TVShowSearchResult>(`/tv/popular?page=${page}`)
}

export async function getTVDetails(id: number): Promise<TVShow> {
  return request<TVShow>(`/tv/${id}`)
}

export async function getSeasonDetails(tvId: number, seasonNumber: number): Promise<Season> {
  return request<Season>(`/tv/${tvId}/season/${seasonNumber}`)
}

// --- Unified Search ---

export async function searchMulti(query: string, page = 1): Promise<MediaSearchResult> {
  return request<MediaSearchResult>(`/search?q=${encodeURIComponent(query)}&page=${page}`)
}

export async function getTrendingAll(): Promise<MediaItem[]> {
  return request<MediaItem[]>('/trending')
}

// --- Torrents ---

export async function searchTorrents(
  tmdbId: number,
  title: string,
  year: number,
  imdbId?: string,
): Promise<TorrentResult[]> {
  const params = new URLSearchParams({
    tmdb_id: String(tmdbId),
    title,
    year: String(year),
  })
  if (imdbId) params.set('imdb_id', imdbId)
  const data = await request<{ results: TorrentResult[] }>(`/torrents/search?${params}`)
  return data.results || []
}

export async function searchTVTorrents(
  title: string,
  season: number,
  year?: string,
): Promise<TorrentResult[]> {
  const params = new URLSearchParams({ title, season: String(season) })
  if (year) params.set('year', year)
  const data = await request<{ results: TorrentResult[] }>(`/torrents/search/tv?${params}`)
  return data.results || []
}

export async function listTorrentFiles(magnetUri: string): Promise<TorrentFile[]> {
  const data = await request<{ files: TorrentFile[] }>('/torrents/files', {
    method: 'POST',
    body: JSON.stringify({ magnet_uri: magnetUri }),
  })
  return data.files || []
}

// --- Streaming ---

export async function startStream(
  tmdbId: number,
  title: string,
  magnetUri: string,
  fileIndex = -1,
): Promise<StreamSession> {
  return request<StreamSession>('/stream/start', {
    method: 'POST',
    body: JSON.stringify({ tmdb_id: tmdbId, title, magnet_uri: magnetUri, file_index: fileIndex }),
  })
}

export function getStreamUrl(sessionId: string, seekTime?: number, audioTrack?: number): string {
  const params = new URLSearchParams()
  if (seekTime && seekTime > 0) params.set('t', seekTime.toFixed(3))
  if (audioTrack !== undefined && audioTrack >= 0) params.set('audio', String(audioTrack))
  const qs = params.toString()
  return `/api/stream/${sessionId}${qs ? '?' + qs : ''}`
}

export async function getStreamStatus(sessionId: string): Promise<StreamStatus> {
  return request<StreamStatus>(`/stream/${sessionId}/status`)
}

export async function stopStream(sessionId: string): Promise<void> {
  await fetch(`${BASE}/stream/${sessionId}`, { method: 'DELETE' })
}

// --- Subtitles ---

export async function searchSubtitles(imdbId: string, lang = 'en'): Promise<SubtitleResult[]> {
  return request<SubtitleResult[]>(`/subtitles/search?imdb_id=${encodeURIComponent(imdbId)}&lang=${lang}`)
}

export function getSubtitleUrl(fileId: number): string {
  return `/api/subtitles/download/${fileId}`
}

// --- Watch History ---

export async function getHistory(): Promise<WatchHistory[]> {
  return request<WatchHistory[]>('/history')
}

export async function getContinueWatching(): Promise<WatchHistory[]> {
  return request<WatchHistory[]>('/history/continue')
}

export async function updateProgress(
  tmdbId: number,
  data: {
    title: string
    poster_path: string
    year: number
    duration: number
    progress: number
    completed: boolean
    quality: string
    magnet_uri: string
  },
): Promise<void> {
  await request(`/history/${tmdbId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function deleteHistory(tmdbId: number): Promise<void> {
  await fetch(`${BASE}/history/${tmdbId}`, { method: 'DELETE' })
}
