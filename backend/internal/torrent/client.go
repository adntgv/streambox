package torrent

import (
	"fmt"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
)

// TorrentClient wraps the anacrolix/torrent client for BitTorrent operations.
type TorrentClient struct {
	client  *torrent.Client
	dataDir string
}

// NewClient creates a new torrent client that stores data in dataDir.
func NewClient(dataDir string) (*TorrentClient, error) {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = dataDir
	cfg.DefaultStorage = storage.NewFileByInfoHash(dataDir)
	cfg.ListenPort = 6881
	cfg.Seed = false
	cfg.EstablishedConnsPerTorrent = 80
	cfg.NoDHT = false
	cfg.DisableTrackers = false
	cfg.HeaderObfuscationPolicy = torrent.HeaderObfuscationPolicy{
		Preferred:        true,
		RequirePreferred: false,
	}

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create torrent client: %w", err)
	}

	return &TorrentClient{
		client:  client,
		dataDir: dataDir,
	}, nil
}

// AddMagnet adds a magnet URI and waits for the torrent metadata to be retrieved.
func (tc *TorrentClient) AddMagnet(magnetURI string) (*torrent.Torrent, error) {
	t, err := tc.client.AddMagnet(magnetURI)
	if err != nil {
		return nil, fmt.Errorf("add magnet: %w", err)
	}
	<-t.GotInfo()
	return t, nil
}

// Close shuts down the torrent client.
func (tc *TorrentClient) Close() {
	tc.client.Close()
}
