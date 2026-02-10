import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import type { MediaItem, PopularItem, WatchHistory } from '../types'
import { getTrendingAll, getPopularTV, getContinueWatching, getPopularHDRezka } from '../api/client'
import MediaGrid from '../components/media/MediaGrid'
import PopularCard from '../components/media/PopularCard'

function ContinueWatchingCard({ item }: { item: WatchHistory }) {
  const posterUrl = item.poster_path
    ? `https://image.tmdb.org/t/p/w342${item.poster_path}`
    : null

  return (
    <Link
      to={`/movie/${item.tmdb_id}`}
      className="group block rounded-lg overflow-hidden shadow-lg bg-zinc-900 transition-transform duration-200 hover:scale-105 hover:shadow-2xl"
    >
      <div className="relative aspect-[2/3] bg-zinc-800">
        {posterUrl ? (
          <img src={posterUrl} alt={item.title} className="w-full h-full object-cover" loading="lazy" />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-zinc-600 text-sm">No Poster</div>
        )}
        {item.quality && (
          <div className="absolute top-2 right-2 bg-black/70 text-indigo-400 text-xs font-semibold px-2 py-0.5 rounded">
            {item.quality}
          </div>
        )}
        {/* Progress bar overlay */}
        <div className="absolute bottom-0 left-0 right-0 h-1 bg-zinc-700">
          <div className="h-full bg-indigo-500 transition-all" style={{ width: `${Math.min(100, item.progress * 100)}%` }} />
        </div>
      </div>
      <div className="p-2">
        <h3 className="text-sm font-medium text-white truncate group-hover:text-indigo-400 transition-colors">
          {item.title}
        </h3>
        {item.year > 0 && <p className="text-xs text-zinc-400 mt-0.5">{item.year}</p>}
      </div>
    </Link>
  )
}

export default function HomePage() {
  const [trending, setTrending] = useState<MediaItem[]>([])
  const [popularTV, setPopularTV] = useState<MediaItem[]>([])
  const [continueWatching, setContinueWatching] = useState<WatchHistory[]>([])
  const [hdrezkaPopular, setHdrezkaPopular] = useState<PopularItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function load() {
      try {
        setLoading(true)
        const [trendingData, popularTVData] = await Promise.all([
          getTrendingAll(),
          getPopularTV(),
        ])
        setTrending(trendingData)
        setPopularTV((popularTVData.results || []).map(s => ({
          id: s.id,
          media_type: 'tv' as const,
          title: s.name,
          overview: s.overview,
          poster_path: s.poster_path,
          backdrop_path: s.backdrop_path,
          date: s.first_air_date,
          vote_average: s.vote_average,
        })))
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load content')
      } finally {
        setLoading(false)
      }
    }
    load()
    // Load non-critical sections separately
    getContinueWatching()
      .then(data => setContinueWatching(data || []))
      .catch(() => {})
    getPopularHDRezka()
      .then(data => setHdrezkaPopular(data || []))
      .catch(() => {})
  }, [])

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8 space-y-10">
        {[1, 2].map(section => (
          <section key={section}>
            <div className="h-7 w-32 bg-zinc-800 rounded animate-pulse mb-4" />
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="rounded-lg overflow-hidden bg-zinc-900">
                  <div className="aspect-[2/3] bg-zinc-800 animate-pulse" />
                  <div className="p-2 space-y-1.5">
                    <div className="h-4 bg-zinc-800 rounded animate-pulse w-3/4" />
                    <div className="h-3 bg-zinc-800 rounded animate-pulse w-1/3" />
                  </div>
                </div>
              ))}
            </div>
          </section>
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-[60vh]">
        <div className="text-red-400 text-lg">{error}</div>
      </div>
    )
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8 space-y-10">
      {continueWatching.length > 0 && (
        <section>
          <h2 className="text-xl font-semibold text-white mb-4">Continue Watching</h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
            {continueWatching.map(item => (
              <ContinueWatchingCard key={item.tmdb_id} item={item} />
            ))}
          </div>
        </section>
      )}
      {hdrezkaPopular.length > 0 && (
        <section>
          <h2 className="text-xl font-semibold text-white mb-4">Popular on HDRezka</h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
            {hdrezkaPopular.map((item, i) => (
              <PopularCard key={i} item={item} />
            ))}
          </div>
        </section>
      )}
      {trending.length > 0 && <MediaGrid items={trending} title="Trending" />}
      {popularTV.length > 0 && <MediaGrid items={popularTV} title="Popular TV Shows" />}
    </div>
  )
}
