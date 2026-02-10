import { Outlet } from 'react-router-dom'
import Header from './Header'

export default function Layout() {
  return (
    <div className="min-h-screen bg-zinc-950 text-white">
      <Header />
      <main className="pt-16">
        <Outlet />
      </main>
    </div>
  )
}
