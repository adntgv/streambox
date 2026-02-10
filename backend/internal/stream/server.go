package stream

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/streambox/backend/internal/torrent"
)

// Server handles HTTP video streaming from torrent sessions.
type Server struct {
	manager *torrent.Manager
}

func NewServer(manager *torrent.Manager) *Server {
	return &Server{manager: manager}
}

// ServeStream serves the video data for a streaming session.
// For MP4/WebM it serves directly via http.ServeContent (Range support).
// For MKV/AVI it pipes through FFmpeg for remuxing to fragmented MP4.
// Supports ?t=<seconds> for time-based seeking on transcoded streams.
func (s *Server) ServeStream(c *gin.Context, sessionID string) {
	sess := s.manager.GetSession(sessionID)
	if sess == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	if !sess.NeedsTranscode {
		// Direct serving — create a fresh reader per request so concurrent
		// Range requests don't conflict on seek position.
		reader := sess.NewReader()
		defer reader.Close()
		http.ServeContent(c.Writer, c.Request, sess.FilePath, time.Time{}, reader.(io.ReadSeeker))
		return
	}

	// Transcoding path — pipe through FFmpeg
	seekTime := 0.0
	if t := c.Query("t"); t != "" {
		if parsed, err := strconv.ParseFloat(t, 64); err == nil && parsed > 0 {
			seekTime = parsed
		}
	}

	audioTrack := -1
	if a := c.Query("audio"); a != "" {
		if parsed, err := strconv.Atoi(a); err == nil && parsed >= 0 {
			audioTrack = parsed
		}
	}

	s.serveTranscoded(c, sess, seekTime, audioTrack)
}

// serveTranscoded pipes the torrent data through FFmpeg to convert MKV/AVI to
// fragmented MP4 that browsers can play. Supports time-based seeking.
func (s *Server) serveTranscoded(c *gin.Context, sess *torrent.Session, seekTime float64, audioTrack int) {
	// Create a fresh reader for this request
	var reader io.Reader
	if seekTime > 0 && sess.Duration > 0 {
		// Approximate byte position based on time ratio
		ratio := seekTime / sess.Duration
		bytePos := int64(ratio * float64(sess.FileSize))
		// Back up 5MB to ensure we hit a keyframe
		if bytePos > 5*1024*1024 {
			bytePos -= 5 * 1024 * 1024
		} else {
			bytePos = 0
		}
		r, err := sess.NewReaderAt(bytePos)
		if err != nil {
			log.Error().Err(err).Float64("seek", seekTime).Msg("failed to seek reader")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "seek failed"})
			return
		}
		defer r.Close()
		reader = r
	} else {
		r := sess.NewReader()
		defer r.Close()
		reader = r
	}

	args := []string{}
	if seekTime > 0 {
		args = append(args, "-ss", strconv.FormatFloat(seekTime, 'f', 3, 64))
	}
	args = append(args, "-i", "pipe:0")
	if audioTrack >= 0 {
		args = append(args, "-map", "0:v:0", "-map", fmt.Sprintf("0:a:%d", audioTrack))
	}
	args = append(args,
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "192k",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof",
		"-f", "mp4",
		"-y",
		"pipe:1",
	)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdin = reader
	cmd.Stdout = c.Writer

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	c.Writer.Header().Set("Content-Type", "video/mp4")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("Cache-Control", "no-cache")

	if err := cmd.Start(); err != nil {
		log.Error().Err(err).Msg("failed to start ffmpeg")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transcoding failed to start"})
		return
	}

	err := cmd.Wait()
	if err != nil {
		if !strings.Contains(stderrBuf.String(), "Broken pipe") &&
			!strings.Contains(err.Error(), "signal: killed") {
			log.Warn().Err(err).Str("stderr", stderrBuf.String()).Msg("ffmpeg exited with error")
		}
	}
}
