import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import type { Movie, TorrentResult } from '../types'
import { getMovieDetails, searchTorrents, startStream } from '../api/client'

export default function MoviePage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [movie, setMovie] = useState<Movie | null>(null)
  const [torrents, setTorrents] = useState<TorrentResult[]>([])
  const [loading, setLoading] = useState(true)
  const [torrentsLoading, setTorrentsLoading] = useState(false)
  const [streamLoading, setStreamLoading] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    async function load() {
      try {
        setLoading(true)
        const data = await getMovieDetails(Number(id))
        setMovie(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load movie')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [id])

  useEffect(() => {
    if (!movie) return
    const year = movie.release_date ? new Date(movie.release_date).getFullYear() : 0

    async function loadTorrents() {
      try {
        setTorrentsLoading(true)
        const data = await searchTorrents(movie!.id, movie!.title, year, movie!.imdb_id)
        setTorrents(data || [])
      } catch {
        // Torrent search failure is non-critical
      } finally {
        setTorrentsLoading(false)
      }
    }
    loadTorrents()
  }, [movie])

  const handleStream = async (torrent: TorrentResult) => {
    if (!movie) return
    try {
      setStreamLoading(torrent.magnet_uri)
      const session = await startStream(movie.id, movie.title, torrent.magnet_uri)
      const year = movie.release_date ? new Date(movie.release_date).getFullYear() : 0
      navigate(`/watch/${session.session_id}`, {
        state: {
          session,
          movieMeta: {
            tmdb_id: movie.id,
            title: movie.title,
            poster_path: movie.poster_path,
            year,
            imdb_id: movie.imdb_id,
            quality: torrent.quality,
            magnet_uri: torrent.magnet_uri,
          },
        },
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start stream')
      setStreamLoading(null)
    }
  }

  if (loading) {
    return (
      <div>
        <div className="h-[50vh] bg-zinc-900 animate-pulse" />
        <div className="max-w-7xl mx-auto px-4 -mt-32 relative z-10 pb-12">
          <div className="flex flex-col md:flex-row gap-8">
            <div className="flex-shrink-0 w-48 md:w-56">
              <div className="aspect-[2/3] bg-zinc-800 rounded-lg animate-pulse" />
            </div>
            <div className="flex-1 space-y-4 pt-4">
              <div className="h-8 bg-zinc-800 rounded animate-pulse w-2/3" />
              <div className="h-4 bg-zinc-800 rounded animate-pulse w-1/3" />
              <div className="flex gap-2">
                {[1, 2, 3].map(i => (
                  <div key={i} className="h-6 w-16 bg-zinc-800 rounded animate-pulse" />
                ))}
              </div>
              <div className="space-y-2">
                <div className="h-4 bg-zinc-800 rounded animate-pulse w-full" />
                <div className="h-4 bg-zinc-800 rounded animate-pulse w-5/6" />
                <div className="h-4 bg-zinc-800 rounded animate-pulse w-4/6" />
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (error || !movie) {
    return (
      <div className="flex items-center justify-center h-[60vh]">
        <div className="text-red-400 text-lg">{error || 'Movie not found'}</div>
      </div>
    )
  }

  const year = movie.release_date ? new Date(movie.release_date).getFullYear() : null
  const backdropUrl = movie.backdrop_path
    ? `https://image.tmdb.org/t/p/w1280${movie.backdrop_path}`
    : null
  const posterUrl = movie.poster_path
    ? `https://image.tmdb.org/t/p/w342${movie.poster_path}`
    : null

  return (
    <div>
      {/* Backdrop */}
      {backdropUrl && (
        <div className="relative h-[50vh] overflow-hidden">
          <img
            src={backdropUrl}
            alt=""
            className="w-full h-full object-cover"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-zinc-950 via-zinc-950/60 to-transparent" />
        </div>
      )}

      <div className="max-w-7xl mx-auto px-4 -mt-32 relative z-10 pb-12">
        <div className="flex flex-col md:flex-row gap-8">
          {/* Poster */}
          {posterUrl && (
            <div className="flex-shrink-0 w-48 md:w-56">
              <img
                src={posterUrl}
                alt={movie.title}
                className="w-full rounded-lg shadow-2xl"
              />
            </div>
          )}

          {/* Details */}
          <div className="flex-1 min-w-0">
            <h1 className="text-3xl md:text-4xl font-bold text-white">
              {movie.title}
            </h1>
            <div className="flex items-center gap-3 mt-2 text-zinc-400 text-sm">
              {year && <span>{year}</span>}
              {movie.runtime > 0 && <span>{movie.runtime} min</span>}
              {movie.vote_average > 0 && (
                <span className="text-yellow-400 font-medium">
                  {movie.vote_average.toFixed(1)}
                </span>
              )}
            </div>

            {movie.genres && movie.genres.length > 0 && (
              <div className="flex flex-wrap gap-2 mt-3">
                {movie.genres.map((g) => (
                  <span
                    key={g.id}
                    className="px-2 py-0.5 text-xs rounded bg-zinc-800 text-zinc-300"
                  >
                    {g.name}
                  </span>
                ))}
              </div>
            )}

            {movie.overview && (
              <p className="mt-4 text-zinc-300 leading-relaxed max-w-2xl">
                {movie.overview}
              </p>
            )}
          </div>
        </div>

        {/* Torrents */}
        <div className="mt-10">
          <h2 className="text-xl font-semibold text-white mb-4">Available Sources</h2>

          {torrentsLoading && (
            <p className="text-zinc-400">Searching for sources...</p>
          )}

          {!torrentsLoading && torrents.length === 0 && (
            <p className="text-zinc-500">No sources found.</p>
          )}

          {torrents.length > 0 && (
            <div className="space-y-2">
              {torrents.map((t, i) => (
                <div
                  key={i}
                  className="flex items-center justify-between gap-4 p-3 rounded-lg bg-zinc-900 border border-zinc-800"
                >
                  <div className="min-w-0 flex-1">
                    <p className="text-sm text-white truncate">{t.title}</p>
                    <div className="flex items-center gap-3 text-xs text-zinc-400 mt-1">
                      {t.quality && <span className="text-indigo-400 font-medium">{t.quality}</span>}
                      {t.size_human && <span>{t.size_human}</span>}
                      <span>S:{t.seeds} P:{t.peers}</span>
                      {t.provider && <span className="text-zinc-500">{t.provider}</span>}
                    </div>
                  </div>
                  <button
                    onClick={() => handleStream(t)}
                    disabled={streamLoading === t.magnet_uri}
                    className="flex-shrink-0 px-4 py-2 rounded bg-indigo-600 text-white text-sm font-medium hover:bg-indigo-500 disabled:opacity-50 transition-colors"
                  >
                    {streamLoading === t.magnet_uri ? 'Starting...' : 'Play'}
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
