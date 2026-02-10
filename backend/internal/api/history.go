package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// getHistory handles GET /api/history
func (s *Server) getHistory(c *gin.Context) {
	history, err := s.db.GetHistory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get watch history", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// getContinueWatching handles GET /api/history/continue
func (s *Server) getContinueWatching(c *gin.Context) {
	items, err := s.db.GetContinueWatching()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get continue watching", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

type updateProgressRequest struct {
	Progress   float64 `json:"progress"`
	Duration   int     `json:"duration"`
	Quality    string  `json:"quality"`
	MagnetURI  string  `json:"magnet_uri"`
	PosterPath string  `json:"poster_path"`
	Title      string  `json:"title"`
	Year       int     `json:"year"`
}

// updateProgress handles PUT /api/history/:tmdb_id
func (s *Server) updateProgress(c *gin.Context) {
	tmdbIDStr := c.Param("tmdb_id")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tmdb_id"})
		return
	}

	var req updateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	if err := s.db.UpsertProgress(tmdbID, req.Title, req.PosterPath, req.Year, req.Duration, req.Progress, req.Quality, req.MagnetURI); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update progress", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "progress updated"})
}

// deleteHistory handles DELETE /api/history/:tmdb_id
func (s *Server) deleteHistory(c *gin.Context) {
	tmdbIDStr := c.Param("tmdb_id")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tmdb_id"})
		return
	}

	if err := s.db.DeleteHistory(tmdbID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete history", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "history entry deleted"})
}
