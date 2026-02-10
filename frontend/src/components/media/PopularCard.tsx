import { useNavigate } from 'react-router-dom'
import type { PopularItem } from '../../types'

interface PopularCardProps {
  item: PopularItem
}

export default function PopularCard({ item }: PopularCardProps) {
  const navigate = useNavigate()

  const handleClick = () => {
    // Extract just the main title (before " / " separator if present)
    let searchTitle = item.title
    const slashIdx = searchTitle.indexOf(' / ')
    if (slashIdx > 0) {
      searchTitle = searchTitle.slice(0, slashIdx)
    }
    navigate(`/search?q=${encodeURIComponent(searchTitle)}`)
  }

  return (
    <button
      onClick={handleClick}
      className="group block rounded-lg overflow-hidden shadow-lg bg-zinc-900 transition-transform duration-200 hover:scale-105 hover:shadow-2xl text-left w-full"
    >
      <div className="relative aspect-[2/3] bg-zinc-800">
        {item.poster ? (
          <img
            src={item.poster}
            alt={item.title}
            className="w-full h-full object-cover"
            loading="lazy"
            referrerPolicy="no-referrer"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-zinc-600 text-sm">
            No Poster
          </div>
        )}
      </div>

      <div className="p-2">
        <h3 className="text-sm font-medium text-white truncate group-hover:text-indigo-400 transition-colors">
          {item.title}
        </h3>
        {item.info && (
          <p className="text-xs text-zinc-400 mt-0.5 truncate">{item.info}</p>
        )}
      </div>
    </button>
  )
}
