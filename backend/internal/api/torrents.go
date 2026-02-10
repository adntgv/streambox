package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// searchTorrents handles GET /api/torrents/search?tmdb_id={id}&title={title}&year={year}&imdb_id={imdb}
func (s *Server) searchTorrents(c *gin.Context) {
	title := c.Query("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'title' is required"})
		return
	}

	imdbID := c.Query("imdb_id")
	year := c.Query("year")

	results, err := s.providers.Search(title, imdbID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search torrents", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}
