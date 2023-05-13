package backend

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

func (b *Backend) UpdateHandler(user *common.User, c *fiber.Ctx) error {
	site := c.Params("sitename")

	integration, ok := b.Integrations[site]
	if !ok {
		return fmt.Errorf("integration does not exist: %s", site)
	}

	account, err := b.LookupAccount(user.ID, site)
	if err != nil {
		return fmt.Errorf("user %s has no account on site: %s", user.Username, site)
	}

	sources, err := integration.GetRawPoints(account)
	if err != nil {
		return fmt.Errorf("problem getting point sources: %w", err)
	}

	if err = b.UpdatePoints(user.ID, site, sources); err != nil {
		return fmt.Errorf("error updating user's points: %w", err)
	}

	return c.Redirect("/")
}

func (b *Backend) UpdatePoints(userID int, siteName string, sources map[string]int) error {
	tx, err := b.Storage.Conn().Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Retrieve the site ID.
	var siteID int
	err = tx.QueryRow(`SELECT id FROM sites WHERE title = ?`, siteName).Scan(&siteID)
	if err != nil {
		return err
	}

	// Retrieve the point sources for the site.
	rows, err := tx.Query(`SELECT id, name FROM point_sources WHERE site_id = ?`, siteID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Map point source names to their IDs.
	sourceIDs := make(map[string]int)
	for rows.Next() {
		var id int
		var name string
		if err = rows.Scan(&id, &name); err != nil {
			return err
		}
		sourceIDs[name] = id
	}

	// Update the points for each source.
	for sourceName, points := range sources {
		sourceID, ok := sourceIDs[sourceName]
		if !ok {
			return fmt.Errorf("source not found: %s", sourceName)
		}

		realPoints := CalculateRealPoints(points, b.Sources[sourceID])

		// Check if the user already has points for this source.
		var currentPoints int
		err = tx.QueryRow(`
			SELECT point_total
			FROM raw_points
			WHERE user_id = ? AND point_source_id = ?
		`, userID, sourceID).Scan(&currentPoints)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		if err == sql.ErrNoRows {
			// If the user doesn't have points for this source yet, insert a new row.
			_, err = tx.Exec(`
				INSERT INTO raw_points (user_id, point_source_id, point_total, real_points, last_updated_date)
				VALUES (?, ?, ?, ?)
			`, userID, sourceID, points, realPoints, time.Now().Unix())
			if err != nil {
				return err
			}
		} else {
			// If the user already has points for this source, update the row.
			_, err = tx.Exec(`
				UPDATE raw_points SET point_total = ?, real_points = ?, last_updated_date = ?
				WHERE user_id = ? AND point_source_id = ?
			`, points, realPoints, time.Now().Unix(), userID, sourceID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *Backend) GetRawPoints(userID int) (total int, sources []*common.AccountPoints, err error) {
	// Step 1: Get all accounts related to the user
	rows, err := b.Storage.Conn().Query(`
		SELECT site_id, username, profile_url
		FROM accounts
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var (
		accountPoints = []*common.AccountPoints{}
		allSitesTotal = 0
	)

	// Step 2: For each account, get all related points
	for rows.Next() {
		account := &common.Account{}
		var siteID int
		if err := rows.Scan(&siteID, &account.Username, &account.ProfileURL); err != nil {
			return 0, nil, err
		}

		// Only include the account if the associated site exists
		site, ok := b.Sites[siteID]
		if !ok {
			continue
		}
		account.Site = site

		accountSources, err := b.getSourcesForAccount(userID, account)
		if err != nil {
			return 0, nil, err
		}

		total := 0
		for _, source := range accountSources {
			total += source.Real
			allSitesTotal += source.Real
		}

		accountPoints = append(accountPoints, &common.AccountPoints{
			Account: account,
			Points:  accountSources,
			Total:   total,
		})
	}

	if err := rows.Err(); err != nil {
		return 0, nil, err
	}

	return allSitesTotal, accountPoints, nil
}

func (b *Backend) getSourcesForAccount(userID int, account *common.Account) ([]*common.Points, error) {
	rows, err := b.Storage.Conn().Query(`
		SELECT point_source_id, point_total, real_points, last_updated_date
		FROM raw_points
		INNER JOIN point_sources ON point_sources.id = raw_points.point_source_id
		WHERE user_id = ? AND point_sources.site_id = ?
	`, userID, account.Site.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := []*common.Points{}
	for rows.Next() {
		point := &common.Points{}
		var sourceID int
		if err := rows.Scan(&sourceID, &point.Raw, &point.Real, &point.LastUpdated); err != nil {
			return nil, err
		}

		// Only include the point if the associated point source exists
		source, ok := b.Sources[sourceID]
		if !ok {
			continue
		}
		point.Source = source

		points = append(points, point)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return points, nil
}

func CalculateRealPoints(raw int, source *common.PointSource) int {
	var low, medium, high int

	if raw > source.MediumUpper {
		// high
		low = source.LowUpper
		medium = source.MediumUpper - source.LowUpper
		high = raw - source.MediumUpper
	} else if raw > source.LowUpper {
		// medium
		low = source.LowUpper
		medium = raw - source.LowUpper
		high = 0
	} else {
		// low
		low = raw
		medium = 0
		high = 0
	}

	return int(math.Round(float64(low)*source.LowRate +
		float64(medium)*source.MediumRate +
		float64(high)*source.HighRate))
}
