package torrent

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	atorrent "github.com/anacrolix/torrent"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambox/backend/internal/db"
	"github.com/streambox/backend/internal/models"
)

// Session holds the runtime state of a single streaming session.
type Session struct {
	models.StreamSession
	torrent        *atorrent.Torrent
	file           *atorrent.File
	reader         atorrent.Reader
	lastBytes      int64
	lastSpeedCheck time.Time
	lastSpeed      int64
}

// GetReader returns the torrent file reader (implements io.Reader and io.ReadSeeker).
func (s *Session) GetReader() atorrent.Reader {
	return s.reader
}

// NewReader creates a fresh reader for concurrent access (e.g. Range requests).
func (s *Session) NewReader() atorrent.Reader {
	r := s.file.NewReader()
	r.SetReadahead(16 * 1024 * 1024)
	r.SetResponsive()
	return r
}

// NewReaderAt creates a reader seeked to the given byte offset.
func (s *Session) NewReaderAt(offset int64) (atorrent.Reader, error) {
	r := s.NewReader()
	if offset > 0 {
		_, err := r.Seek(offset, io.SeekStart)
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("seek to %d: %w", offset, err)
		}
	}
	return r, nil
}

// Manager manages active torrent streaming sessions.
type Manager struct {
	client   *TorrentClient
	db       *db.DB
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewManager(client *TorrentClient, database *db.DB) *Manager {
	return &Manager{
		client:   client,
		db:       database,
		sessions: make(map[string]*Session),
	}
}

// StartStream adds a magnet URI to the torrent client, identifies the largest
// video file, creates a reader, and returns a StreamSession.
func (m *Manager) StartStream(tmdbID int, title, magnetURI string) (*models.StreamSession, error) {
	log.Info().Str("title", title).Msg("starting stream")

	t, err := m.client.AddMagnet(magnetURI)
	if err != nil {
		return nil, fmt.Errorf("add magnet: %w", err)
	}

	videoFile := findLargestVideoFile(t.Files())
	if videoFile == nil {
		t.Drop()
		return nil, fmt.Errorf("no video file found in torrent")
	}

	reader := videoFile.NewReader()
	reader.SetReadahead(16 * 1024 * 1024)
	reader.SetResponsive()

	contentType := detectContentType(videoFile.DisplayPath())
	needsTranscode := needsTranscoding(videoFile.DisplayPath())

	sess := &Session{
		StreamSession: models.StreamSession{
			ID:             uuid.New().String(),
			TMDbID:         tmdbID,
			Title:          title,
			MagnetURI:      magnetURI,
			InfoHash:       t.InfoHash().HexString(),
			FilePath:       videoFile.DisplayPath(),
			FileSize:       videoFile.Length(),
			ContentType:    contentType,
			NeedsTranscode: needsTranscode,
			Status:         "ready",
		},
		torrent: t,
		file:    videoFile,
		reader:  reader,
	}

	m.mu.Lock()
	m.sessions[sess.ID] = sess
	m.mu.Unlock()

	// Probe duration and audio tracks in background
	go m.probeMedia(sess)

	log.Info().
		Str("session_id", sess.ID).
		Str("file", videoFile.DisplayPath()).
		Int64("size", videoFile.Length()).
		Bool("transcode", needsTranscode).
		Msg("stream session created")

	return &sess.StreamSession, nil
}

// probeMedia runs ffprobe on the torrent data to extract duration and audio tracks.
func (m *Manager) probeMedia(sess *Session) {
	r := sess.file.NewReader()
	r.SetReadahead(10 * 1024 * 1024)
	r.SetResponsive()
	defer r.Close()

	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		"-select_streams", "a",
		"-analyzeduration", "5000000",
		"-probesize", "10000000",
		"-i", "pipe:0",
	)
	cmd.Stdin = r

	out, err := cmd.Output()
	if err != nil {
		log.Warn().Err(err).Str("session", sess.ID).Msg("ffprobe failed")
		return
	}

	var probe struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
		Streams []struct {
			Index     int    `json:"index"`
			CodecType string `json:"codec_type"`
			Tags      struct {
				Language string `json:"language"`
				Title    string `json:"title"`
			} `json:"tags"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &probe); err != nil {
		log.Warn().Err(err).Msg("parse ffprobe output")
		return
	}

	// Parse duration
	dur, err := strconv.ParseFloat(probe.Format.Duration, 64)
	if err != nil {
		log.Warn().Err(err).Str("raw", probe.Format.Duration).Msg("parse duration")
	}

	// Parse audio tracks
	var tracks []models.AudioTrack
	for i, s := range probe.Streams {
		title := s.Tags.Title
		if title == "" {
			lang := s.Tags.Language
			if lang == "" {
				lang = "und"
			}
			title = fmt.Sprintf("Track %d (%s)", i+1, lang)
		}
		tracks = append(tracks, models.AudioTrack{
			Index:    i,
			Language: s.Tags.Language,
			Title:    title,
		})
	}

	m.mu.Lock()
	if dur > 0 {
		sess.Duration = dur
	}
	sess.AudioTracks = tracks
	m.mu.Unlock()

	log.Info().
		Str("session_id", sess.ID).
		Float64("duration_sec", dur).
		Int("audio_tracks", len(tracks)).
		Msg("probed media info")
}

func formatDuration(seconds float64) string {
	h := int(seconds) / 3600
	min := (int(seconds) % 3600) / 60
	sec := int(seconds) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, min, sec)
	}
	return fmt.Sprintf("%d:%02d", min, sec)
}

// GetSession returns the runtime Session by ID (used by stream server).
func (m *Manager) GetSession(id string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

// GetStatus returns download/buffering status for a session.
func (m *Manager) GetStatus(sessionID string) (*models.StreamStatus, error) {
	m.mu.RLock()
	sess := m.sessions[sessionID]
	m.mu.RUnlock()

	if sess == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	t := sess.torrent
	stats := t.Stats()
	bytesCompleted := sess.file.BytesCompleted()

	// Dynamic readahead based on conditions
	downloadPct := float64(bytesCompleted) / float64(sess.FileSize) * 100
	var readahead int64 = 16 * 1024 * 1024
	if stats.ActivePeers < 3 {
		readahead = 64 * 1024 * 1024
	} else if downloadPct < 10 {
		readahead = 32 * 1024 * 1024
	}
	sess.reader.SetReadahead(readahead)

	// Calculate download speed
	now := time.Now()
	var speed int64
	if !sess.lastSpeedCheck.IsZero() {
		elapsed := now.Sub(sess.lastSpeedCheck).Seconds()
		if elapsed > 0 {
			speed = int64(float64(bytesCompleted-sess.lastBytes) / elapsed)
			if speed < 0 {
				speed = 0
			}
			sess.lastSpeed = speed
		}
	}
	sess.lastBytes = bytesCompleted
	sess.lastSpeedCheck = now

	return &models.StreamStatus{
		Status:          sess.Status,
		DownloadedBytes: bytesCompleted,
		TotalBytes:      sess.FileSize,
		DownloadSpeed:   speed,
		PeersConnected:  stats.ActivePeers,
		BufferedPercent: float64(bytesCompleted) / float64(sess.FileSize) * 100,
		Duration:        sess.Duration,
		AudioTracks:     sess.AudioTracks,
	}, nil
}

// StopSession stops and removes a streaming session.
func (m *Manager) StopSession(sessionID string) error {
	m.mu.Lock()
	sess := m.sessions[sessionID]
	if sess == nil {
		m.mu.Unlock()
		return fmt.Errorf("session not found: %s", sessionID)
	}
	delete(m.sessions, sessionID)
	m.mu.Unlock()

	if sess.reader != nil {
		sess.reader.Close()
	}
	sess.torrent.Drop()

	log.Info().Str("session_id", sessionID).Msg("stream session stopped")
	return nil
}

// findLargestVideoFile finds the largest file with a video extension in the torrent.
func findLargestVideoFile(files []*atorrent.File) *atorrent.File {
	videoExts := map[string]bool{
		".mp4": true, ".mkv": true, ".avi": true, ".webm": true,
		".mov": true, ".wmv": true, ".flv": true, ".m4v": true,
	}

	var largest *atorrent.File
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.DisplayPath()))
		if !videoExts[ext] {
			continue
		}
		if largest == nil || f.Length() > largest.Length() {
			largest = f
		}
	}
	return largest
}

// needsTranscoding returns true if the file format is not natively playable in browsers.
func needsTranscoding(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext != ".mp4" && ext != ".webm"
}

// detectContentType returns the MIME type for a video file.
func detectContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp4", ".m4v":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mkv":
		return "video/x-matroska"
	case ".avi":
		return "video/x-msvideo"
	default:
		return "application/octet-stream"
	}
}
