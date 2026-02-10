package api

import (
	"fmt"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/streambox/backend/internal/config"
	"github.com/streambox/backend/internal/db"
	"github.com/streambox/backend/internal/tmdb"
	"github.com/streambox/backend/internal/torrent"
	"github.com/streambox/backend/internal/stream"
	"github.com/streambox/backend/internal/subtitle"
)

type Server struct {
	config         *config.Config
	router         *gin.Engine
	tmdb           *tmdb.Client
	providers      *torrent.ProviderRegistry
	torrentMgr     *torrent.Manager
	streamSrv      *stream.Server
	subtitleClient *subtitle.Client
	db             *db.DB
}

func NewServer(cfg *config.Config, database *db.DB, tmdbClient *tmdb.Client, providers *torrent.ProviderRegistry, torrentMgr *torrent.Manager, streamSrv *stream.Server, subClient *subtitle.Client) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost:")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	s := &Server{
		config:         cfg,
		router:         r,
		tmdb:           tmdbClient,
		providers:      providers,
		torrentMgr:     torrentMgr,
		streamSrv:      streamSrv,
		subtitleClient: subClient,
		db:             database,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		// Movies (TMDB proxy)
		api.GET("/movies/search", s.searchMovies)
		api.GET("/movies/trending", s.getTrending)
		api.GET("/movies/popular", s.getPopular)
		api.GET("/movies/:id", s.getMovieDetails)

		// Torrents
		api.GET("/torrents/search", s.searchTorrents)

		// Streaming
		api.POST("/stream/start", s.startStream)
		api.GET("/stream/:id", s.serveStream)
		api.GET("/stream/:id/status", s.getStreamStatus)
		api.DELETE("/stream/:id", s.stopStream)

		// Subtitles
		api.GET("/subtitles/search", s.searchSubtitles)
		api.GET("/subtitles/download/:id", s.downloadSubtitle)

		// Watch History
		api.GET("/history", s.getHistory)
		api.GET("/history/continue", s.getContinueWatching)
		api.PUT("/history/:tmdb_id", s.updateProgress)
		api.POST("/history/:tmdb_id", s.updateProgress) // sendBeacon can only POST
		api.DELETE("/history/:tmdb_id", s.deleteHistory)
	}

	// Serve React SPA static files
	s.router.Static("/assets", "./static/assets")
	s.router.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.File("./static/index.html")
		}
	})
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	return s.router.Run(addr)
}
