package torrent

import (
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/streambox/backend/internal/models"
)

// Provider is the interface that torrent search providers must implement.
type Provider interface {
	Name() string
	Search(title, imdbID string, year string) ([]models.TorrentResult, error)
}

// ProviderRegistry holds all registered torrent search providers and
// searches them concurrently.
type ProviderRegistry struct {
	providers []Provider
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{}
}

func (r *ProviderRegistry) Register(p Provider) {
	r.providers = append(r.providers, p)
}

// Search queries all registered providers concurrently and returns
// aggregated results.
func (r *ProviderRegistry) Search(title, imdbID string, year string) ([]models.TorrentResult, error) {
	var (
		allResults []models.TorrentResult
		mu         sync.Mutex
		wg         sync.WaitGroup
	)

	for _, p := range r.providers {
		wg.Add(1)
		go func(prov Provider) {
			defer wg.Done()
			results, err := prov.Search(title, imdbID, year)
			if err != nil {
				log.Warn().Err(err).Str("provider", prov.Name()).Msg("torrent search failed")
				return
			}
			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()
		}(p)
	}

	wg.Wait()
	return allResults, nil
}
