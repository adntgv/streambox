package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// searchTV handles GET /api/tv/search?q={query}&page={page}
func (s *Server) searchTV(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	results, err := s.tmdb.SearchTV(query, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search tv shows", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// getTrendingTV handles GET /api/tv/trending
func (s *Server) getTrendingTV(c *gin.Context) {
	results, err := s.tmdb.GetTrendingTV()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get trending tv shows", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// getPopularTV handles GET /api/tv/popular?page={page}
func (s *Server) getPopularTV(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	results, err := s.tmdb.GetPopularTV(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get popular tv shows", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// getTVDetails handles GET /api/tv/:id
func (s *Server) getTVDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid TV show ID"})
		return
	}

	show, err := s.tmdb.GetTVDetails(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tv show details", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, show)
}

// getSeasonDetails handles GET /api/tv/:id/season/:season
func (s *Server) getSeasonDetails(c *gin.Context) {
	tvID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid TV show ID"})
		return
	}

	seasonNum, err := strconv.Atoi(c.Param("season"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid season number"})
		return
	}

	season, err := s.tmdb.GetSeasonDetails(tvID, seasonNum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get season details", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, season)
}
