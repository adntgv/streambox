import type { MediaItem } from '../../types'
import MediaCard from './MediaCard'

interface MediaGridProps {
  items: MediaItem[]
  title?: string
}

export default function MediaGrid({ items, title }: MediaGridProps) {
  return (
    <section>
      {title && (
        <h2 className="text-xl font-semibold text-white mb-4">{title}</h2>
      )}
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
        {items.map((item) => (
          <MediaCard key={`${item.media_type}-${item.id}`} item={item} />
        ))}
      </div>
    </section>
  )
}
