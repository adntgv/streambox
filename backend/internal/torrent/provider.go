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

// TVSearcher is an optional interface for providers that support TV series search.
type TVSearcher interface {
	SearchTV(title string, seasonNum int, year string) ([]models.TorrentResult, error)
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

// SearchTV queries providers that implement TVSearcher concurrently.
func (r *ProviderRegistry) SearchTV(title string, seasonNum int, year string) ([]models.TorrentResult, error) {
	var (
		allResults []models.TorrentResult
		mu         sync.Mutex
		wg         sync.WaitGroup
	)

	for _, p := range r.providers {
		tvp, ok := p.(TVSearcher)
		if !ok {
			continue
		}
		wg.Add(1)
		go func(prov TVSearcher, name string) {
			defer wg.Done()
			results, err := prov.SearchTV(title, seasonNum, year)
			if err != nil {
				log.Warn().Err(err).Str("provider", name).Msg("tv torrent search failed")
				return
			}
			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()
		}(tvp, p.Name())
	}

	wg.Wait()
	return allResults, nil
}
