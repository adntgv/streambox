import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import type { Movie } from '../types'
import { searchMovies } from '../api/client'
import MovieGrid from '../components/movie/MovieGrid'

export default function SearchPage() {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q') || ''
  const [movies, setMovies] = useState<Movie[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(0)

  useEffect(() => {
    if (!query) return
    setPage(1)
    setMovies([])
  }, [query])

  useEffect(() => {
    if (!query) return

    async function load() {
      try {
        setLoading(true)
        setError(null)
        const data = await searchMovies(query, page)
        setMovies(data.results || [])
        setTotalPages(data.total_pages)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Search failed')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [query, page])

  if (!query) {
    return (
      <div className="flex items-center justify-center h-[60vh]">
        <p className="text-zinc-400 text-lg">Enter a search query to find movies</p>
      </div>
    )
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-semibold text-white mb-6">
        Results for "{query}"
      </h1>

      {loading && (
        <div className="flex items-center justify-center h-40">
          <div className="text-zinc-400">Searching...</div>
        </div>
      )}

      {error && (
        <div className="text-red-400 mb-4">{error}</div>
      )}

      {!loading && movies.length === 0 && (
        <p className="text-zinc-400">No results found.</p>
      )}

      {movies.length > 0 && <MovieGrid movies={movies} />}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-4 mt-8">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page <= 1}
            className="px-4 py-2 rounded bg-zinc-800 text-white disabled:opacity-40 hover:bg-zinc-700 transition-colors"
          >
            Previous
          </button>
          <span className="text-zinc-400 text-sm">
            Page {page} of {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page >= totalPages}
            className="px-4 py-2 rounded bg-zinc-800 text-white disabled:opacity-40 hover:bg-zinc-700 transition-colors"
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
