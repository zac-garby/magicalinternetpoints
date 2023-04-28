package backend

import (
	"fmt"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
	"github.com/zac-garby/magicalinternetpoints/lib/integrations"
)

type Backend struct {
	Storage  *sqlite3.Storage
	Sessions *session.Store

	Sites   map[int]*common.Site
	Sources map[int]*common.PointSource

	Integrations map[string]integrations.Integration
}

func New(storage *sqlite3.Storage) (*Backend, error) {
	sessions := session.New(session.Config{
		Storage: storage,
	})

	backend := &Backend{
		Storage:  storage,
		Sessions: sessions,

		Integrations: map[string]integrations.Integration{
			"GitHub": &integrations.GitHub{},
			"Reddit": &integrations.Reddit{},
		},
	}

	if err := backend.LoadInitData(); err != nil {
		return nil, err
	}

	return backend, nil
}

func (b *Backend) LoadInitData() error {
	// Query to get all sites
	sitesQuery := `
        SELECT id, title, url, score_description
        FROM sites
    `
	sitesRows, err := b.Storage.Conn().Query(sitesQuery)
	if err != nil {
		return fmt.Errorf("failed to query sites: %v", err)
	}
	defer sitesRows.Close()

	// Build map of site IDs to site objects
	b.Sites = make(map[int]*common.Site)
	for sitesRows.Next() {
		var site common.Site
		if err := sitesRows.Scan(&site.ID, &site.Title, &site.URL, &site.ScoreDescription); err != nil {
			return fmt.Errorf("failed to scan site row: %v", err)
		}

		site.Sources = make([]*common.PointSource, 0)
		b.Sites[site.ID] = &site
	}
	if err := sitesRows.Err(); err != nil {
		return fmt.Errorf("failed to iterate over site rows: %v", err)
	}

	// Query to get all point sources
	sourcesQuery := `
        SELECT
			ps.id, ps.name, ps.description, ps.site_id,
			ps.low_upper, ps.medium_upper,
			ps.low_rate, ps.medium_rate, ps.high_rate
        FROM point_sources AS ps
    `
	sourcesRows, err := b.Storage.Conn().Query(sourcesQuery)
	if err != nil {
		return fmt.Errorf("failed to query point sources: %v", err)
	}
	defer sourcesRows.Close()

	// Build map of point source IDs to point source objects
	b.Sources = make(map[int]*common.PointSource)
	for sourcesRows.Next() {
		var source common.PointSource
		var siteID int
		if err := sourcesRows.Scan(
			&source.ID, &source.Name, &source.Description, &siteID,
			&source.LowUpper, &source.MediumUpper,
			&source.LowRate, &source.MediumRate, &source.HighRate); err != nil {
			return fmt.Errorf("failed to scan point source row: %v", err)
		}
		site, ok := b.Sites[siteID]
		if !ok {
			return fmt.Errorf("invalid site ID for point source %d: %d", source.ID, siteID)
		}
		source.Site = site
		site.Sources = append(site.Sources, &source)
		b.Sources[source.ID] = &source
	}
	if err := sourcesRows.Err(); err != nil {
		return fmt.Errorf("failed to iterate over point source rows: %v", err)
	}

	return nil
}
