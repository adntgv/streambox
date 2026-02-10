package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/streambox/backend/internal/api"
	"github.com/streambox/backend/internal/config"
	"github.com/streambox/backend/internal/db"
	"github.com/streambox/backend/internal/hdrezka"
	"github.com/streambox/backend/internal/stream"
	"github.com/streambox/backend/internal/subtitle"
	"github.com/streambox/backend/internal/tmdb"
	"github.com/streambox/backend/internal/torrent"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	if err := os.MkdirAll(cfg.TorrentDir, 0755); err != nil {
		log.Fatal().Err(err).Msg("failed to create torrent directory")
	}

	database, err := db.New(cfg.DBPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
	}
	defer database.Close()

	tmdbClient := tmdb.NewClient(cfg.TMDBAPIKey)

	torrentClient, err := torrent.NewClient(cfg.TorrentDir)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize torrent client")
	}
	defer torrentClient.Close()

	providers := torrent.NewProviderRegistry()
	if cfg.RutrackerUsername != "" && cfg.RutrackerPassword != "" {
		rt := torrent.NewRutracker(cfg.RutrackerMirror, cfg.RutrackerUsername, cfg.RutrackerPassword)
		providers.Register(rt)
		log.Info().Msg("rutracker provider registered")
	}
	providers.Register(torrent.NewYTS())

	torrentMgr := torrent.NewManager(torrentClient, database)
	streamSrv := stream.NewServer(torrentMgr)

	var subClient *subtitle.Client
	if cfg.OpenSubtitlesKey != "" {
		subClient = subtitle.NewClient(cfg.OpenSubtitlesKey)
	}

	hdrezkaClient := hdrezka.NewClient()

	server := api.NewServer(cfg, database, tmdbClient, providers, torrentMgr, streamSrv, subClient, hdrezkaClient)

	log.Info().Int("port", cfg.Port).Msg("starting StreamBox server")
	if err := server.Run(); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}
