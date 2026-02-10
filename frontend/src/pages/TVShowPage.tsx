import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import type { TVShow, Season, Episode, TorrentResult, TorrentFile } from '../types'
import { getTVDetails, getSeasonDetails, searchTVTorrents, listTorrentFiles, startStream } from '../api/client'

export default function TVShowPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [show, setShow] = useState<TVShow | null>(null)
  const [selectedSeason, setSelectedSeason] = useState<number>(1)
  const [seasonDetail, setSeasonDetail] = useState<Season | null>(null)
  const [torrents, setTorrents] = useState<TorrentResult[]>([])
  const [loading, setLoading] = useState(true)
  const [seasonLoading, setSeasonLoading] = useState(false)
  const [torrentsLoading, setTorrentsLoading] = useState(false)
  const [streamLoading, setStreamLoading] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  // File picker state for season packs
  const [filePicker, setFilePicker] = useState<{ torrent: TorrentResult; files: TorrentFile[] } | null>(null)
  const [filesLoading, setFilesLoading] = useState(false)

  // Load show details
  useEffect(() => {
    if (!id) return
    async function load() {
      try {
        setLoading(true)
        const data = await getTVDetails(Number(id))
        setShow(data)
        // Select first non-specials season
        const realSeasons = (data.seasons || []).filter(s => s.season_number > 0)
        if (realSeasons.length > 0) {
          setSelectedSeason(realSeasons[0].season_number)
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load TV show')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [id])

  // Load season episodes when season changes
  useEffect(() => {
    if (!show) return
    async function loadSeason() {
      try {
        setSeasonLoading(true)
        const data = await getSeasonDetails(show!.id, selectedSeason)
        setSeasonDetail(data)
      } catch {
        setSeasonDetail(null)
      } finally {
        setSeasonLoading(false)
      }
    }
    loadSeason()
  }, [show, selectedSeason])

  // Search torrents when show or season changes
  useEffect(() => {
    if (!show) return
    async function loadTorrents() {
      try {
        setTorrentsLoading(true)
        const year = show!.first_air_date ? show!.first_air_date.slice(0, 4) : undefined
        const data = await searchTVTorrents(show!.name, selectedSeason, year)
        setTorrents(data || [])
      } catch {
        setTorrents([])
      } finally {
        setTorrentsLoading(false)
      }
    }
    loadTorrents()
  }, [show, selectedSeason])

  const handleStream = async (torrent: TorrentResult, fileIndex = -1) => {
    if (!show) return
    try {
      setStreamLoading(torrent.magnet_uri)
      const session = await startStream(show.id, show.name, torrent.magnet_uri, fileIndex)
      const year = show.first_air_date ? new Date(show.first_air_date).getFullYear() : 0
      navigate(`/watch/${session.session_id}`, {
        state: {
          session,
          movieMeta: {
            tmdb_id: show.id,
            title: show.name,
            poster_path: show.poster_path,
            year,
            quality: torrent.quality,
            magnet_uri: torrent.magnet_uri,
            media_type: 'tv',
            season_number: selectedSeason,
          },
        },
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start stream')
      setStreamLoading(null)
    }
  }

  const handlePickFiles = async (torrent: TorrentResult) => {
    try {
      setFilesLoading(true)
      setStreamLoading(torrent.magnet_uri)
      const files = await listTorrentFiles(torrent.magnet_uri)
      if (files.length <= 1) {
        // Single file or no files — just stream directly
        handleStream(torrent)
        return
      }
      setFilePicker({ torrent, files })
    } catch {
      // Fallback: stream without file selection
      handleStream(torrent)
    } finally {
      setFilesLoading(false)
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
              <div className="space-y-2">
                <div className="h-4 bg-zinc-800 rounded animate-pulse w-full" />
                <div className="h-4 bg-zinc-800 rounded animate-pulse w-5/6" />
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (error || !show) {
    return (
      <div className="flex items-center justify-center h-[60vh]">
        <div className="text-red-400 text-lg">{error || 'TV show not found'}</div>
      </div>
    )
  }

  const year = show.first_air_date ? new Date(show.first_air_date).getFullYear() : null
  const backdropUrl = show.backdrop_path
    ? `https://image.tmdb.org/t/p/w1280${show.backdrop_path}`
    : null
  const posterUrl = show.poster_path
    ? `https://image.tmdb.org/t/p/w342${show.poster_path}`
    : null
  const realSeasons = (show.seasons || []).filter(s => s.season_number > 0)

  return (
    <div>
      {/* Backdrop */}
      {backdropUrl && (
        <div className="relative h-[50vh] overflow-hidden">
          <img src={backdropUrl} alt="" className="w-full h-full object-cover" />
          <div className="absolute inset-0 bg-gradient-to-t from-zinc-950 via-zinc-950/60 to-transparent" />
        </div>
      )}

      <div className="max-w-7xl mx-auto px-4 -mt-32 relative z-10 pb-12">
        <div className="flex flex-col md:flex-row gap-8">
          {posterUrl && (
            <div className="flex-shrink-0 w-48 md:w-56">
              <img src={posterUrl} alt={show.name} className="w-full rounded-lg shadow-2xl" />
            </div>
          )}

          <div className="flex-1 min-w-0">
            <h1 className="text-3xl md:text-4xl font-bold text-white">{show.name}</h1>
            <div className="flex items-center gap-3 mt-2 text-zinc-400 text-sm">
              {year && <span>{year}</span>}
              {show.number_of_seasons && <span>{show.number_of_seasons} seasons</span>}
              {show.vote_average > 0 && (
                <span className="text-yellow-400 font-medium">
                  {show.vote_average.toFixed(1)}
                </span>
              )}
            </div>

            {show.genres && show.genres.length > 0 && (
              <div className="flex flex-wrap gap-2 mt-3">
                {show.genres.map(g => (
                  <span key={g.id} className="px-2 py-0.5 text-xs rounded bg-zinc-800 text-zinc-300">
                    {g.name}
                  </span>
                ))}
              </div>
            )}

            {show.overview && (
              <p className="mt-4 text-zinc-300 leading-relaxed max-w-2xl">{show.overview}</p>
            )}
          </div>
        </div>

        {/* Season tabs */}
        {realSeasons.length > 0 && (
          <div className="mt-8">
            <div className="flex gap-2 overflow-x-auto pb-2">
              {realSeasons.map(s => (
                <button
                  key={s.season_number}
                  onClick={() => { setSelectedSeason(s.season_number); setFilePicker(null) }}
                  className={`px-4 py-2 rounded-lg text-sm font-medium whitespace-nowrap transition-colors ${
                    selectedSeason === s.season_number
                      ? 'bg-indigo-600 text-white'
                      : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700 hover:text-white'
                  }`}
                >
                  Season {s.season_number}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Episodes */}
        <div className="mt-6">
          {seasonLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="h-20 bg-zinc-900 rounded-lg animate-pulse" />
              ))}
            </div>
          ) : seasonDetail?.episodes && seasonDetail.episodes.length > 0 ? (
            <div className="space-y-2">
              {seasonDetail.episodes.map((ep: Episode) => (
                <div key={ep.id} className="flex gap-4 p-3 rounded-lg bg-zinc-900 border border-zinc-800">
                  {ep.still_path ? (
                    <img
                      src={`https://image.tmdb.org/t/p/w300${ep.still_path}`}
                      alt={ep.name}
                      className="w-32 h-18 object-cover rounded flex-shrink-0"
                      loading="lazy"
                    />
                  ) : (
                    <div className="w-32 h-18 bg-zinc-800 rounded flex-shrink-0 flex items-center justify-center text-zinc-600 text-xs">
                      E{ep.episode_number}
                    </div>
                  )}
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-white">
                      {ep.episode_number}. {ep.name}
                    </p>
                    {ep.overview && (
                      <p className="text-xs text-zinc-400 mt-1 line-clamp-2">{ep.overview}</p>
                    )}
                    <div className="flex gap-3 text-xs text-zinc-500 mt-1">
                      {ep.runtime > 0 && <span>{ep.runtime} min</span>}
                      {ep.vote_average > 0 && <span className="text-yellow-400">{ep.vote_average.toFixed(1)}</span>}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : null}
        </div>

        {/* Sources */}
        <div className="mt-10">
          <h2 className="text-xl font-semibold text-white mb-4">Available Sources</h2>

          {torrentsLoading && <p className="text-zinc-400">Searching for sources...</p>}

          {!torrentsLoading && torrents.length === 0 && (
            <p className="text-zinc-500">No sources found for Season {selectedSeason}.</p>
          )}

          {torrents.length > 0 && (
            <div className="space-y-2">
              {torrents.map((t, i) => (
                <div key={i} className="flex items-center justify-between gap-4 p-3 rounded-lg bg-zinc-900 border border-zinc-800">
                  <div className="min-w-0 flex-1">
                    <p className="text-sm text-white truncate">{t.title}</p>
                    <div className="flex items-center gap-3 text-xs text-zinc-400 mt-1">
                      {t.quality && <span className="text-indigo-400 font-medium">{t.quality}</span>}
                      {t.size_human && <span>{t.size_human}</span>}
                      <span>S:{t.seeds} P:{t.peers}</span>
                    </div>
                  </div>
                  <button
                    onClick={() => handlePickFiles(t)}
                    disabled={streamLoading === t.magnet_uri}
                    className="flex-shrink-0 px-4 py-2 rounded bg-indigo-600 text-white text-sm font-medium hover:bg-indigo-500 disabled:opacity-50 transition-colors"
                  >
                    {streamLoading === t.magnet_uri ? (filesLoading ? 'Loading...' : 'Starting...') : 'Play'}
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* File picker overlay */}
        {filePicker && (
          <div className="fixed inset-0 bg-black/80 z-50 flex items-center justify-center p-4" onClick={() => setFilePicker(null)}>
            <div className="bg-zinc-900 rounded-xl max-w-2xl w-full max-h-[80vh] overflow-hidden" onClick={e => e.stopPropagation()}>
              <div className="p-4 border-b border-zinc-800 flex items-center justify-between">
                <h3 className="text-lg font-semibold text-white">Select Episode</h3>
                <button onClick={() => setFilePicker(null)} className="text-zinc-400 hover:text-white">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div className="overflow-y-auto max-h-[60vh] p-2">
                {filePicker.files.map(f => {
                  const fileName = f.path.split('/').pop() || f.path
                  return (
                    <button
                      key={f.index}
                      onClick={() => { setFilePicker(null); handleStream(filePicker.torrent, f.index) }}
                      className="w-full text-left p-3 rounded-lg hover:bg-zinc-800 transition-colors"
                    >
                      <p className="text-sm text-white truncate">{fileName}</p>
                      <p className="text-xs text-zinc-400 mt-0.5">{f.size_human}</p>
                    </button>
                  )
                })}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
