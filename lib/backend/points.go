package backend

import (
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

func (b *Backend) GetRawPoints(userID int) ([]*common.AccountPoints, error) {
	// Step 1: Get all accounts related to the user
	rows, err := b.Storage.Conn().Query(`
		SELECT site_id, username, profile_url
		FROM accounts
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accountPoints := []*common.AccountPoints{}

	// Step 2: For each account, get all related points
	for rows.Next() {
		account := &common.Account{}
		var siteID int
		if err := rows.Scan(&siteID, &account.Username, &account.ProfileURL); err != nil {
			return nil, err
		}

		// Only include the account if the associated site exists
		site, ok := b.Sites[siteID]
		if !ok {
			continue
		}
		account.Site = site

		accountSources, err := b.getSourcesForAccount(userID, account)
		if err != nil {
			return nil, err
		}

		accountPoints = append(accountPoints, &common.AccountPoints{
			Account: account,
			Points:  accountSources,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accountPoints, nil
}

func (b *Backend) getSourcesForAccount(userID int, account *common.Account) ([]*common.Points, error) {
	rows, err := b.Storage.Conn().Query(`
		SELECT point_source_id, point_total, last_updated_date
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
		if err := rows.Scan(&sourceID, &point.Raw, &point.LastUpdated); err != nil {
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
