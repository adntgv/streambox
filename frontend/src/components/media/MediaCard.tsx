import { Link } from 'react-router-dom'
import type { MediaItem } from '../../types'

interface MediaCardProps {
  item: MediaItem
}

export default function MediaCard({ item }: MediaCardProps) {
  const year = item.date ? new Date(item.date).getFullYear() : null
  const rating = item.vote_average ? item.vote_average.toFixed(1) : null
  const posterUrl = item.poster_path
    ? `https://image.tmdb.org/t/p/w342${item.poster_path}`
    : null
  const link = item.media_type === 'tv' ? `/tv/${item.id}` : `/movie/${item.id}`

  return (
    <Link
      to={link}
      className="group block rounded-lg overflow-hidden shadow-lg bg-zinc-900 transition-transform duration-200 hover:scale-105 hover:shadow-2xl"
    >
      <div className="relative aspect-[2/3] bg-zinc-800">
        {posterUrl ? (
          <img
            src={posterUrl}
            alt={item.title}
            className="w-full h-full object-cover"
            loading="lazy"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-zinc-600 text-sm">
            No Poster
          </div>
        )}

        {rating && (
          <div className="absolute top-2 right-2 bg-black/70 text-yellow-400 text-xs font-semibold px-2 py-0.5 rounded">
            {rating}
          </div>
        )}

        {item.media_type === 'tv' && (
          <div className="absolute top-2 left-2 bg-indigo-600/90 text-white text-[10px] font-bold px-1.5 py-0.5 rounded">
            TV
          </div>
        )}
      </div>

      <div className="p-2">
        <h3 className="text-sm font-medium text-white truncate group-hover:text-indigo-400 transition-colors">
          {item.title}
        </h3>
        {year && (
          <p className="text-xs text-zinc-400 mt-0.5">{year}</p>
        )}
      </div>
    </Link>
  )
}
