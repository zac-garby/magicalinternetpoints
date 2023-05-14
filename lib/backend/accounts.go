package backend

import (
	"fmt"

	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

func (b *Backend) GetAccounts(userID int) ([]*common.Account, error) {
	rows, err := b.Storage.Conn().Query(`
	SELECT site_id, username, profile_url
	FROM accounts
	WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*common.Account{}

	for rows.Next() {
		var (
			account common.Account
			siteID  int
		)

		if err := rows.Scan(&siteID, &account.Username, &account.ProfileURL); err != nil {
			return nil, err
		}

		site, ok := b.Sites[siteID]
		if !ok {
			return nil, fmt.Errorf("site does not exist with ID: %d", siteID)
		}

		account.Site = site

		accounts = append(accounts, &account)
	}

	return accounts, nil
}

func (b *Backend) GetNonLinkedSites(userID int) ([]*common.Site, error) {
	rows, err := b.Storage.Conn().Query(`
	SELECT id
	FROM sites
	WHERE id NOT IN (
		SELECT site_id
		FROM accounts
		WHERE user_id = ?
	)
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sites := []*common.Site{}

	for rows.Next() {
		var siteID int

		if err := rows.Scan(&siteID); err != nil {
			return nil, err
		}

		site, ok := b.Sites[siteID]
		if !ok {
			return nil, fmt.Errorf("site does not exist with ID: %d", siteID)
		}

		sites = append(sites, site)
	}

	return sites, nil
}
