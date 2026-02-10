import { useEffect, useState, useRef, useCallback } from 'react'
import { useParams, useNavigate, useLocation } from 'react-router-dom'
import { getStreamUrl, getStreamStatus, stopStream, updateProgress, searchSubtitles, getSubtitleUrl } from '../api/client'
import type { StreamSession, StreamStatus, SubtitleResult, AudioTrack } from '../types'

interface MovieMeta {
  tmdb_id: number
  title: string
  poster_path: string
  year: number
  imdb_id?: string
  quality?: string
  magnet_uri?: string
}

const SPEED_OPTIONS = [0.5, 0.75, 1, 1.25, 1.5, 2]

export default function PlayerPage() {
  const { sessionId } = useParams<{ sessionId: string }>()
  const navigate = useNavigate()
  const location = useLocation()
  const locState = location.state as { session?: StreamSession; movieMeta?: MovieMeta } | null
  const session = locState?.session
  const movieMeta = locState?.movieMeta

  const videoRef = useRef<HTMLVideoElement>(null)
  const progressRef = useRef<HTMLDivElement>(null)
  const hideTimeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined)
  const progressSaveRef = useRef<ReturnType<typeof setInterval>>(undefined)
  const lastSavedProgressRef = useRef(0)

  const [status, setStatus] = useState<StreamStatus | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [ready, setReady] = useState(false)
  const [playing, setPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [videoDuration, setVideoDuration] = useState(0)
  const [bufferedEnd, setBufferedEnd] = useState(0)
  const [volume, setVolume] = useState(1)
  const [muted, setMuted] = useState(false)
  const [showControls, setShowControls] = useState(true)
  const [seekPreview, setSeekPreview] = useState<number | null>(null)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [buffering, setBuffering] = useState(false)
  const [playbackSpeed, setPlaybackSpeed] = useState(1)
  const [isPiP, setIsPiP] = useState(false)
  const [retryCount, setRetryCount] = useState(0)
  const [showSpeedMenu, setShowSpeedMenu] = useState(false)
  const [subtitles, setSubtitles] = useState<{ lang: string; results: SubtitleResult[] }[]>([])
  const [showSubMenu, setShowSubMenu] = useState(false)
  const [activeSubtitle, setActiveSubtitle] = useState<number | null>(null)
  const [doubleTapSide, setDoubleTapSide] = useState<'left' | 'right' | null>(null)
  const [audioTracks, setAudioTracks] = useState<AudioTrack[]>([])
  const [selectedAudio, setSelectedAudio] = useState(-1)
  const [showAudioMenu, setShowAudioMenu] = useState(false)

  const needsTranscode = session?.needs_transcode ?? false
  const backendDuration = status?.duration ?? session?.duration ?? 0
  const totalDuration = videoDuration > 0 && isFinite(videoDuration) ? videoDuration : backendDuration

  // --- Progress saving ---
  const saveProgress = useCallback(() => {
    if (!movieMeta || !totalDuration || totalDuration <= 0) return
    const progress = currentTime / totalDuration
    if (Math.abs(progress - lastSavedProgressRef.current) < 0.01) return
    lastSavedProgressRef.current = progress
    updateProgress(movieMeta.tmdb_id, {
      title: movieMeta.title,
      poster_path: movieMeta.poster_path,
      year: movieMeta.year,
      duration: Math.floor(totalDuration),
      progress,
      completed: progress > 0.9,
      quality: movieMeta.quality || '',
      magnet_uri: movieMeta.magnet_uri || '',
    }).catch(() => {})
  }, [movieMeta, currentTime, totalDuration])

  // Save every 10 seconds
  useEffect(() => {
    if (!ready) return
    progressSaveRef.current = setInterval(saveProgress, 10000)
    return () => { if (progressSaveRef.current) clearInterval(progressSaveRef.current) }
  }, [ready, saveProgress])

  // Save on pause
  useEffect(() => {
    if (!playing && ready && currentTime > 0) saveProgress()
  }, [playing])

  // Save on unload
  useEffect(() => {
    const onUnload = () => {
      if (!movieMeta || !totalDuration) return
      const progress = currentTime / totalDuration
      const body = JSON.stringify({
        title: movieMeta.title,
        poster_path: movieMeta.poster_path,
        year: movieMeta.year,
        duration: Math.floor(totalDuration),
        progress,
        completed: progress > 0.9,
        quality: movieMeta.quality || '',
        magnet_uri: movieMeta.magnet_uri || '',
      })
      navigator.sendBeacon(`/api/history/${movieMeta.tmdb_id}`, new Blob([body], { type: 'application/json' }))
    }
    window.addEventListener('beforeunload', onUnload)
    return () => window.removeEventListener('beforeunload', onUnload)
  }, [movieMeta, currentTime, totalDuration])

  // --- Subtitle loading ---
  useEffect(() => {
    if (!movieMeta?.imdb_id) return
    const loadSubs = async () => {
      const results: { lang: string; results: SubtitleResult[] }[] = []
      for (const lang of ['ru', 'en']) {
        try {
          const subs = await searchSubtitles(movieMeta.imdb_id!, lang)
          if (subs.length > 0) results.push({ lang, results: subs })
        } catch { /* ignore */ }
      }
      setSubtitles(results)
    }
    loadSubs()
  }, [movieMeta?.imdb_id])

  // --- Poll stream status ---
  useEffect(() => {
    if (!sessionId) return
    let cancelled = false
    const poll = async () => {
      try {
        const s = await getStreamStatus(sessionId)
        if (cancelled) return
        setStatus(s)
        if (s.audio_tracks && s.audio_tracks.length > 0 && audioTracks.length === 0) {
          setAudioTracks(s.audio_tracks)
        }
        if (s.status === 'ready' || s.buffered_percent >= 2) setReady(true)
      } catch { /* ignore */ }
    }
    poll()
    const interval = setInterval(poll, 3000)
    return () => { cancelled = true; clearInterval(interval) }
  }, [sessionId])

  // --- Video event listeners ---
  useEffect(() => {
    const video = videoRef.current
    if (!video) return

    const onTimeUpdate = () => {
      setCurrentTime(video.currentTime)
      if (video.buffered.length > 0) setBufferedEnd(video.buffered.end(video.buffered.length - 1))
    }
    const onDurationChange = () => {
      if (video.duration && isFinite(video.duration)) setVideoDuration(video.duration)
    }
    const onPlay = () => { setPlaying(true); setBuffering(false) }
    const onPause = () => setPlaying(false)
    const onEnded = () => setPlaying(false)
    const onWaiting = () => setBuffering(true)
    const onPlaying = () => setBuffering(false)
    const onEnterPiP = () => setIsPiP(true)
    const onLeavePiP = () => setIsPiP(false)

    video.addEventListener('timeupdate', onTimeUpdate)
    video.addEventListener('durationchange', onDurationChange)
    video.addEventListener('loadedmetadata', onDurationChange)
    video.addEventListener('play', onPlay)
    video.addEventListener('pause', onPause)
    video.addEventListener('ended', onEnded)
    video.addEventListener('waiting', onWaiting)
    video.addEventListener('playing', onPlaying)
    video.addEventListener('enterpictureinpicture', onEnterPiP)
    video.addEventListener('leavepictureinpicture', onLeavePiP)

    return () => {
      video.removeEventListener('timeupdate', onTimeUpdate)
      video.removeEventListener('durationchange', onDurationChange)
      video.removeEventListener('loadedmetadata', onDurationChange)
      video.removeEventListener('play', onPlay)
      video.removeEventListener('pause', onPause)
      video.removeEventListener('ended', onEnded)
      video.removeEventListener('waiting', onWaiting)
      video.removeEventListener('playing', onPlaying)
      video.removeEventListener('enterpictureinpicture', onEnterPiP)
      video.removeEventListener('leavepictureinpicture', onLeavePiP)
    }
  }, [ready])

  // --- Auto-hide controls ---
  const showControlsTemporarily = useCallback(() => {
    setShowControls(true)
    setShowSpeedMenu(false)
    setShowSubMenu(false)
    setShowAudioMenu(false)
    if (hideTimeoutRef.current) clearTimeout(hideTimeoutRef.current)
    hideTimeoutRef.current = setTimeout(() => {
      if (playing) setShowControls(false)
    }, 3000)
  }, [playing])

  useEffect(() => {
    if (!playing) {
      setShowControls(true)
      if (hideTimeoutRef.current) clearTimeout(hideTimeoutRef.current)
    }
  }, [playing])

  // --- Fullscreen ---
  useEffect(() => {
    const onFSChange = () => setIsFullscreen(!!document.fullscreenElement)
    document.addEventListener('fullscreenchange', onFSChange)
    return () => document.removeEventListener('fullscreenchange', onFSChange)
  }, [])

  // --- Seeking ---
  const handleSeek = useCallback((time: number) => {
    const video = videoRef.current
    if (!video || !sessionId) return
    const clampedTime = Math.max(0, Math.min(time, totalDuration || Infinity))
    if (needsTranscode) {
      video.src = getStreamUrl(sessionId, clampedTime, selectedAudio >= 0 ? selectedAudio : undefined)
      video.play().catch(() => {})
      setCurrentTime(clampedTime)
    } else {
      video.currentTime = clampedTime
    }
  }, [sessionId, needsTranscode, totalDuration, selectedAudio])

  // --- Error recovery ---
  const handleVideoError = useCallback(() => {
    if (retryCount >= 3) {
      setError('Playback failed after multiple retries')
      return
    }
    const delay = Math.pow(2, retryCount) * 1000
    setRetryCount(r => r + 1)
    setBuffering(true)
    setTimeout(() => {
      const video = videoRef.current
      if (!video || !sessionId) return
      const audio = selectedAudio >= 0 ? selectedAudio : undefined
      if (needsTranscode && currentTime > 0) {
        video.src = getStreamUrl(sessionId, currentTime, audio)
      } else {
        video.src = getStreamUrl(sessionId, undefined, audio)
      }
      video.play().catch(() => {})
    }, delay)
  }, [retryCount, sessionId, needsTranscode, currentTime, selectedAudio])

  // --- Keyboard shortcuts ---
  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      const video = videoRef.current
      if (!video) return
      switch (e.key) {
        case ' ': case 'k':
          e.preventDefault()
          video.paused ? video.play() : video.pause()
          showControlsTemporarily()
          break
        case 'ArrowLeft':
          e.preventDefault()
          handleSeek(currentTime - (e.shiftKey ? 30 : 10))
          showControlsTemporarily()
          break
        case 'ArrowRight':
          e.preventDefault()
          handleSeek(currentTime + (e.shiftKey ? 30 : 10))
          showControlsTemporarily()
          break
        case 'ArrowUp':
          e.preventDefault()
          setVolume(v => { const nv = Math.min(1, v + 0.1); video.volume = nv; return nv })
          showControlsTemporarily()
          break
        case 'ArrowDown':
          e.preventDefault()
          setVolume(v => { const nv = Math.max(0, v - 0.1); video.volume = nv; return nv })
          showControlsTemporarily()
          break
        case 'f': e.preventDefault(); toggleFullscreen(); break
        case 'm': e.preventDefault(); toggleMute(); showControlsTemporarily(); break
        case 'Escape': if (isFullscreen) document.exitFullscreen(); break
        case '<': case ',':
          e.preventDefault()
          cycleSpeed(-1)
          showControlsTemporarily()
          break
        case '>': case '.':
          e.preventDefault()
          cycleSpeed(1)
          showControlsTemporarily()
          break
        case 'c':
          e.preventDefault()
          setShowSubMenu(v => !v)
          showControlsTemporarily()
          break
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [currentTime, playing, isFullscreen, showControlsTemporarily, handleSeek])

  // --- Speed control ---
  const cycleSpeed = useCallback((dir: number) => {
    setPlaybackSpeed(prev => {
      const idx = SPEED_OPTIONS.indexOf(prev)
      const nextIdx = Math.max(0, Math.min(SPEED_OPTIONS.length - 1, idx + dir))
      const next = SPEED_OPTIONS[nextIdx]
      if (videoRef.current) videoRef.current.playbackRate = next
      return next
    })
  }, [])

  // --- Double-tap seek (mobile) ---
  const lastTapRef = useRef<{ time: number; x: number }>({ time: 0, x: 0 })
  const doubleTapTimeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined)

  const handleVideoAreaClick = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    if ((e.target as HTMLElement).closest('.player-controls')) return
    const now = Date.now()
    const gap = now - lastTapRef.current.time
    const rect = e.currentTarget.getBoundingClientRect()
    const x = e.clientX

    if (gap < 300) {
      // Double tap
      if (doubleTapTimeoutRef.current) clearTimeout(doubleTapTimeoutRef.current)
      const isLeft = x < rect.left + rect.width / 2
      handleSeek(currentTime + (isLeft ? -10 : 10))
      setDoubleTapSide(isLeft ? 'left' : 'right')
      setTimeout(() => setDoubleTapSide(null), 500)
    } else {
      // Single tap — delayed to distinguish from double
      doubleTapTimeoutRef.current = setTimeout(() => {
        togglePlay()
        showControlsTemporarily()
      }, 300)
    }
    lastTapRef.current = { time: now, x }
  }, [currentTime, handleSeek, showControlsTemporarily])

  const handleProgressClick = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    if (!totalDuration || !progressRef.current) return
    const rect = progressRef.current.getBoundingClientRect()
    const ratio = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width))
    handleSeek(ratio * totalDuration)
    setSeekPreview(null)
  }, [totalDuration, handleSeek])

  const handleProgressHover = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    if (!totalDuration || !progressRef.current) return
    const rect = progressRef.current.getBoundingClientRect()
    const ratio = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width))
    setSeekPreview(ratio * totalDuration)
  }, [totalDuration])

  const togglePlay = useCallback(() => {
    const v = videoRef.current
    if (v) v.paused ? v.play() : v.pause()
  }, [])

  const toggleMute = useCallback(() => {
    const v = videoRef.current
    if (v) { v.muted = !v.muted; setMuted(!muted) }
  }, [muted])

  const toggleFullscreen = useCallback(() => {
    const c = document.querySelector('.player-container')
    if (!c) return
    document.fullscreenElement ? document.exitFullscreen() : c.requestFullscreen()
  }, [])

  const togglePiP = useCallback(async () => {
    const v = videoRef.current
    if (!v) return
    try {
      if (document.pictureInPictureElement) {
        await document.exitPictureInPicture()
      } else {
        await v.requestPictureInPicture()
      }
    } catch { /* not supported */ }
  }, [])

  const handleBack = useCallback(async () => {
    saveProgress()
    if (sessionId) { try { await stopStream(sessionId) } catch { /* ignore */ } }
    navigate(-1)
  }, [sessionId, navigate, saveProgress])

  const selectSubtitle = useCallback((fileId: number | null) => {
    setActiveSubtitle(fileId)
    setShowSubMenu(false)
    const video = videoRef.current
    if (!video) return
    // Remove existing tracks
    while (video.textTracks.length > 0) {
      const track = video.querySelector('track')
      if (track) track.remove()
      else break
    }
    if (fileId !== null) {
      const track = document.createElement('track')
      track.kind = 'subtitles'
      track.src = getSubtitleUrl(fileId)
      track.default = true
      video.appendChild(track)
      if (video.textTracks[0]) video.textTracks[0].mode = 'showing'
    }
  }, [])

  // --- Format helpers ---
  const formatTime = (s: number) => {
    if (!s || !isFinite(s)) return '0:00'
    const h = Math.floor(s / 3600)
    const m = Math.floor((s % 3600) / 60)
    const sec = Math.floor(s % 60)
    return h > 0
      ? `${h}:${m.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}`
      : `${m}:${sec.toString().padStart(2, '0')}`
  }

  const formatSpeed = (bps: number) => {
    if (bps > 1048576) return `${(bps / 1048576).toFixed(1)} MB/s`
    if (bps > 1024) return `${(bps / 1024).toFixed(0)} KB/s`
    return `${bps} B/s`
  }

  if (!sessionId) {
    return (
      <div className="flex items-center justify-center h-screen bg-black">
        <p className="text-red-400">No session ID</p>
      </div>
    )
  }

  const streamUrl = getStreamUrl(sessionId, undefined, selectedAudio >= 0 ? selectedAudio : undefined)
  const progressPercent = totalDuration > 0 ? (currentTime / totalDuration) * 100 : 0
  const bufferedPercent = totalDuration > 0 ? (bufferedEnd / totalDuration) * 100 : 0
  const downloadPercent = status?.buffered_percent ?? 0
  const canPiP = typeof document !== 'undefined' && 'pictureInPictureEnabled' in document && document.pictureInPictureEnabled

  return (
    <div
      className="player-container fixed inset-0 bg-black z-50 flex flex-col select-none"
      onMouseMove={showControlsTemporarily}
      onClick={handleVideoAreaClick}
      style={{ cursor: showControls ? 'default' : 'none' }}
    >
      {/* Top bar */}
      <div className={`player-controls absolute top-0 left-0 right-0 z-20 p-4 bg-gradient-to-b from-black/80 to-transparent transition-opacity duration-300 ${showControls ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}>
        <div className="flex items-center gap-4">
          <button onClick={(e) => { e.stopPropagation(); handleBack() }} className="text-white hover:text-zinc-300 transition-colors">
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <span className="text-white text-sm font-medium truncate">{movieMeta?.title || session?.title || 'Playing'}</span>
        </div>
      </div>

      {/* Video */}
      <div className="flex-1 flex items-center justify-center relative">
        {error ? (
          <div className="text-center">
            <p className="text-red-400 text-lg mb-4">{error}</p>
            <button onClick={() => { setError(null); setRetryCount(0); handleVideoError() }} className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-500">Try Again</button>
          </div>
        ) : !ready ? (
          <div className="text-center">
            <div className="w-12 h-12 border-4 border-zinc-600 border-t-white rounded-full animate-spin mx-auto mb-4" />
            <div className="text-white text-lg mb-2">Buffering...</div>
            {status && (
              <div className="text-zinc-400 text-sm space-y-1">
                <p>Downloaded: {status.buffered_percent.toFixed(1)}%</p>
                {status.download_speed > 0 && <p>Speed: {formatSpeed(status.download_speed)}</p>}
                <p>Peers: {status.peers_connected}</p>
              </div>
            )}
          </div>
        ) : (
          <>
            <video
              ref={videoRef}
              src={streamUrl}
              className="w-full h-full object-contain"
              autoPlay
              playsInline
              onError={() => handleVideoError()}
            />
            {/* Mid-playback buffering spinner */}
            {buffering && (
              <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <div className="w-14 h-14 border-4 border-zinc-600 border-t-white rounded-full animate-spin" />
              </div>
            )}
            {/* Retry indicator */}
            {retryCount > 0 && retryCount < 3 && buffering && (
              <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <div className="mt-20 text-zinc-400 text-sm">Reconnecting... ({retryCount}/3)</div>
              </div>
            )}
            {/* Double-tap indicator */}
            {doubleTapSide && (
              <div className={`absolute top-1/2 -translate-y-1/2 ${doubleTapSide === 'left' ? 'left-12' : 'right-12'} pointer-events-none`}>
                <div className="bg-black/60 text-white text-lg font-bold px-4 py-2 rounded-full animate-pulse">
                  {doubleTapSide === 'left' ? '-10s' : '+10s'}
                </div>
              </div>
            )}
          </>
        )}
      </div>

      {/* Bottom controls */}
      {ready && (
        <div className={`player-controls absolute bottom-0 left-0 right-0 z-20 bg-gradient-to-t from-black/90 via-black/60 to-transparent pt-16 pb-4 px-4 transition-opacity duration-300 ${showControls ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}>
          {/* Progress bar */}
          <div
            ref={progressRef}
            className="group relative h-1.5 hover:h-3 bg-zinc-700/60 rounded-full cursor-pointer transition-all mb-3"
            onClick={(e) => { e.stopPropagation(); handleProgressClick(e) }}
            onMouseMove={handleProgressHover}
            onMouseLeave={() => setSeekPreview(null)}
          >
            <div className="absolute inset-y-0 left-0 bg-zinc-600/50 rounded-full transition-all" style={{ width: `${Math.min(100, downloadPercent)}%` }} />
            <div className="absolute inset-y-0 left-0 bg-zinc-400/40 rounded-full transition-all" style={{ width: `${Math.min(100, bufferedPercent)}%` }} />
            <div className="absolute inset-y-0 left-0 bg-indigo-500 rounded-full" style={{ width: `${Math.min(100, progressPercent)}%` }} />
            <div className="absolute top-1/2 -translate-y-1/2 w-3.5 h-3.5 bg-indigo-400 rounded-full shadow-lg opacity-0 group-hover:opacity-100 transition-opacity" style={{ left: `calc(${Math.min(100, progressPercent)}% - 7px)` }} />
            {seekPreview !== null && (
              <div className="absolute -top-8 -translate-x-1/2 bg-black/90 text-white text-xs px-2 py-1 rounded pointer-events-none" style={{ left: `${(seekPreview / totalDuration) * 100}%` }}>
                {formatTime(seekPreview)}
              </div>
            )}
          </div>

          {/* Controls row */}
          <div className="flex items-center gap-2">
            {/* Play/Pause */}
            <button onClick={(e) => { e.stopPropagation(); togglePlay() }} className="text-white hover:text-zinc-300 transition-colors">
              {playing ? (
                <svg className="w-7 h-7" fill="currentColor" viewBox="0 0 24 24"><path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z" /></svg>
              ) : (
                <svg className="w-7 h-7" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z" /></svg>
              )}
            </button>

            {/* Skip -10s */}
            <button onClick={(e) => { e.stopPropagation(); handleSeek(currentTime - 10) }} className="text-white/70 hover:text-white transition-colors" title="Rewind 10s">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12.066 11.2a1 1 0 000 1.6l5.334 4A1 1 0 0019 16V8a1 1 0 00-1.6-.8l-5.333 4zM4.066 11.2a1 1 0 000 1.6l5.334 4A1 1 0 0011 16V8a1 1 0 00-1.6-.8l-5.334 4z" />
              </svg>
            </button>

            {/* Skip +10s */}
            <button onClick={(e) => { e.stopPropagation(); handleSeek(currentTime + 10) }} className="text-white/70 hover:text-white transition-colors" title="Forward 10s">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M11.933 12.8a1 1 0 000-1.6L6.6 7.2A1 1 0 005 8v8a1 1 0 001.6.8l5.333-4zM19.933 12.8a1 1 0 000-1.6l-5.333-4A1 1 0 0013 8v8a1 1 0 001.6.8l5.333-4z" />
              </svg>
            </button>

            {/* Volume */}
            <div className="flex items-center gap-1 group/vol">
              <button onClick={(e) => { e.stopPropagation(); toggleMute() }} className="text-white/70 hover:text-white transition-colors">
                {muted || volume === 0 ? (
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z" /></svg>
                ) : volume < 0.5 ? (
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M18.5 12c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM5 9v6h4l5 5V4L9 9H5z" /></svg>
                ) : (
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z" /></svg>
                )}
              </button>
              <input type="range" min={0} max={1} step={0.05} value={muted ? 0 : volume}
                onChange={(e) => { e.stopPropagation(); const v = parseFloat(e.target.value); setVolume(v); setMuted(v === 0); if (videoRef.current) { videoRef.current.volume = v; videoRef.current.muted = v === 0 } }}
                onClick={(e) => e.stopPropagation()}
                className="w-0 group-hover/vol:w-20 transition-all accent-white h-1 cursor-pointer overflow-hidden"
              />
            </div>

            {/* Time */}
            <div className="text-white/80 text-xs font-mono ml-1">
              {formatTime(currentTime)}
              {totalDuration > 0 && <span className="text-zinc-400"> / {formatTime(totalDuration)}</span>}
            </div>

            <div className="flex-1" />

            {/* Torrent stats */}
            {status && (
              <div className="hidden sm:flex items-center gap-3 text-xs text-zinc-400">
                {status.download_speed > 0 && <span>{formatSpeed(status.download_speed)}</span>}
                <span>{status.peers_connected} peers</span>
                <span>{status.buffered_percent.toFixed(0)}%</span>
              </div>
            )}

            {/* Speed */}
            <div className="relative">
              <button onClick={(e) => { e.stopPropagation(); setShowSpeedMenu(v => !v); setShowSubMenu(false) }}
                className="text-white/70 hover:text-white transition-colors text-xs font-medium px-1.5 py-0.5 rounded border border-zinc-600 hover:border-zinc-400">
                {playbackSpeed}x
              </button>
              {showSpeedMenu && (
                <div className="absolute bottom-full right-0 mb-2 bg-zinc-900/95 border border-zinc-700 rounded-lg py-1 min-w-[80px]" onClick={(e) => e.stopPropagation()}>
                  {SPEED_OPTIONS.map(s => (
                    <button key={s} onClick={() => { setPlaybackSpeed(s); if (videoRef.current) videoRef.current.playbackRate = s; setShowSpeedMenu(false) }}
                      className={`block w-full text-left px-3 py-1.5 text-sm transition-colors ${s === playbackSpeed ? 'text-indigo-400 bg-zinc-800' : 'text-white hover:bg-zinc-800'}`}>
                      {s}x
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* Subtitles */}
            {subtitles.length > 0 && (
              <div className="relative">
                <button onClick={(e) => { e.stopPropagation(); setShowSubMenu(v => !v); setShowSpeedMenu(false) }}
                  className={`transition-colors ${activeSubtitle ? 'text-indigo-400' : 'text-white/70 hover:text-white'}`} title="Subtitles">
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
                  </svg>
                </button>
                {showSubMenu && (
                  <div className="absolute bottom-full right-0 mb-2 bg-zinc-900/95 border border-zinc-700 rounded-lg py-1 min-w-[160px]" onClick={(e) => e.stopPropagation()}>
                    <button onClick={() => selectSubtitle(null)}
                      className={`block w-full text-left px-3 py-1.5 text-sm transition-colors ${activeSubtitle === null ? 'text-indigo-400 bg-zinc-800' : 'text-white hover:bg-zinc-800'}`}>
                      Off
                    </button>
                    {subtitles.map(({ lang, results }) => (
                      <button key={lang} onClick={() => selectSubtitle(results[0].file_id)}
                        className={`block w-full text-left px-3 py-1.5 text-sm transition-colors ${activeSubtitle === results[0].file_id ? 'text-indigo-400 bg-zinc-800' : 'text-white hover:bg-zinc-800'}`}>
                        {lang === 'ru' ? 'Russian' : 'English'} ({results[0].name})
                      </button>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* Audio tracks */}
            {audioTracks.length > 1 && needsTranscode && (
              <div className="relative">
                <button onClick={(e) => { e.stopPropagation(); setShowAudioMenu(v => !v); setShowSpeedMenu(false); setShowSubMenu(false) }}
                  className="text-white/70 hover:text-white transition-colors" title="Audio track">
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
                  </svg>
                </button>
                {showAudioMenu && (
                  <div className="absolute bottom-full right-0 mb-2 bg-zinc-900/95 border border-zinc-700 rounded-lg py-1 min-w-[200px] max-h-60 overflow-y-auto" onClick={(e) => e.stopPropagation()}>
                    {audioTracks.map(track => (
                      <button key={track.index} onClick={() => {
                        setSelectedAudio(track.index)
                        setShowAudioMenu(false)
                        const video = videoRef.current
                        if (video && sessionId) {
                          const time = video.currentTime || currentTime
                          video.src = getStreamUrl(sessionId, time > 0 ? time : undefined, track.index)
                          video.play().catch(() => {})
                        }
                      }}
                        className={`block w-full text-left px-3 py-1.5 text-sm transition-colors ${track.index === selectedAudio ? 'text-indigo-400 bg-zinc-800' : 'text-white hover:bg-zinc-800'}`}>
                        {track.title}
                        {track.language && <span className="text-zinc-500 ml-1">({track.language})</span>}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* PiP */}
            {canPiP && (
              <button onClick={(e) => { e.stopPropagation(); togglePiP() }}
                className={`transition-colors ${isPiP ? 'text-indigo-400' : 'text-white/70 hover:text-white'}`} title="Picture in Picture">
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3 7a2 2 0 012-2h14a2 2 0 012 2v10a2 2 0 01-2 2H5a2 2 0 01-2-2V7z" />
                  <path strokeLinecap="round" strokeLinejoin="round" d="M13 13h6v4h-6z" />
                </svg>
              </button>
            )}

            {/* Fullscreen */}
            <button onClick={(e) => { e.stopPropagation(); toggleFullscreen() }} className="text-white/70 hover:text-white transition-colors">
              {isFullscreen ? (
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 9V4.5M9 9H4.5M9 9L3.75 3.75M9 15v4.5M9 15H4.5M9 15l-5.25 5.25M15 9h4.5M15 9V4.5M15 9l5.25-5.25M15 15h4.5M15 15v4.5m0-4.5l5.25 5.25" />
                </svg>
              ) : (
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 3.75v4.5m0-4.5h4.5m-4.5 0L9 9M3.75 20.25v-4.5m0 4.5h4.5m-4.5 0L9 15M20.25 3.75h-4.5m4.5 0v4.5m0-4.5L15 9m5.25 11.25h-4.5m4.5 0v-4.5m0 4.5L15 15" />
                </svg>
              )}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
