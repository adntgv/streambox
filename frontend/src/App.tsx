import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from './components/layout/Layout'
import HomePage from './pages/HomePage'
import SearchPage from './pages/SearchPage'
import MoviePage from './pages/MoviePage'
import PlayerPage from './pages/PlayerPage'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<HomePage />} />
          <Route path="/search" element={<SearchPage />} />
          <Route path="/movie/:id" element={<MoviePage />} />
        </Route>
        {/* PlayerPage is fullscreen, no layout */}
        <Route path="/watch/:sessionId" element={<PlayerPage />} />
      </Routes>
    </BrowserRouter>
  )
}
