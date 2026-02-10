package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type startStreamRequest struct {
	TMDbID   int    `json:"tmdb_id" binding:"required"`
	Title    string `json:"title" binding:"required"`
	MagnetURI string `json:"magnet_uri" binding:"required"`
}

// startStream handles POST /api/stream/start
func (s *Server) startStream(c *gin.Context) {
	var req startStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	session, err := s.torrentMgr.StartStream(req.TMDbID, req.Title, req.MagnetURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start stream", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// serveStream handles GET /api/stream/:id
func (s *Server) serveStream(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	s.streamSrv.ServeStream(c, sessionID)
}

// getStreamStatus handles GET /api/stream/:id/status
func (s *Server) getStreamStatus(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	status, err := s.torrentMgr.GetStatus(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stream status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// stopStream handles DELETE /api/stream/:id
func (s *Server) stopStream(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	if err := s.torrentMgr.StopSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stop stream", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stream stopped"})
}
