import { useState, useCallback } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'

export default function Header() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [query, setQuery] = useState(searchParams.get('q') || '')

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault()
      const trimmed = query.trim()
      if (trimmed) {
        navigate(`/search?q=${encodeURIComponent(trimmed)}`)
      }
    },
    [query, navigate],
  )

  return (
    <header className="fixed top-0 left-0 right-0 z-50 bg-zinc-900/95 backdrop-blur-sm border-b border-zinc-800">
      <div className="max-w-7xl mx-auto px-4 h-16 flex items-center justify-between gap-4">
        <Link to="/" className="flex-shrink-0">
          <h1 className="text-xl font-bold text-white tracking-tight">
            <span className="text-indigo-500">Stream</span>Box
          </h1>
        </Link>

        <form onSubmit={handleSubmit} className="flex-1 max-w-md">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search movies..."
            className="w-full px-4 py-2 rounded-lg bg-zinc-800 text-white placeholder-zinc-400 border border-zinc-700 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors text-sm"
          />
        </form>
      </div>
    </header>
  )
}
