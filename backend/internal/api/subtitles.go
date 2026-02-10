package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// searchSubtitles handles GET /api/subtitles/search?imdb_id={id}&lang={en}
func (s *Server) searchSubtitles(c *gin.Context) {
	if s.subtitleClient == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "subtitles not configured"})
		return
	}

	imdbID := c.Query("imdb_id")
	if imdbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'imdb_id' is required"})
		return
	}

	lang := c.DefaultQuery("lang", "en")

	results, err := s.subtitleClient.Search(imdbID, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search subtitles", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// downloadSubtitle handles GET /api/subtitles/download/:id
func (s *Server) downloadSubtitle(c *gin.Context) {
	if s.subtitleClient == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "subtitles not configured"})
		return
	}

	idStr := c.Param("id")
	fileID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subtitle file ID"})
		return
	}

	data, err := s.subtitleClient.Download(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to download subtitle", "details": err.Error()})
		return
	}

	c.Data(http.StatusOK, "text/vtt", data)
}
