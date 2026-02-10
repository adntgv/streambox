package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// searchMovies handles GET /api/movies/search?q={query}&page={page}
func (s *Server) searchMovies(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	results, err := s.tmdb.Search(query, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search movies", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// getTrending handles GET /api/movies/trending
func (s *Server) getTrending(c *gin.Context) {
	results, err := s.tmdb.GetTrending()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get trending movies", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// getPopular handles GET /api/movies/popular?page={page}
func (s *Server) getPopular(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	results, err := s.tmdb.GetPopular(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get popular movies", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// getMovieDetails handles GET /api/movies/:id
func (s *Server) getMovieDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie ID"})
		return
	}

	movie, err := s.tmdb.GetDetails(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get movie details", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, movie)
}
