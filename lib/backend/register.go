package backend

import (
	"database/sql"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/markbates/goth"
	"github.com/shareed2k/goth_fiber"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

func (b *Backend) RegisterAccountHandler(u *common.User, c *fiber.Ctx) error {
	authUser, err := goth_fiber.CompleteUserAuth(c)
	if err != nil {
		return fmt.Errorf("error authenticating: %w", err)
	}

	if err := goth_fiber.Logout(c); err != nil {
		return fmt.Errorf("error logging out: %w", err)
	}

	if err := b.ensureAccountNotRegistered(u, &authUser); err != nil {
		return fmt.Errorf("could not register account: %w", err)
	}

	if err := b.RegisterAccount(u, &authUser); err != nil {
		return fmt.Errorf("could not register account: %w", err)
	}

	return c.Redirect("/accounts")
}

func (b *Backend) UnlinkHandler(u *common.User, c *fiber.Ctx) error {
	siteName := c.Params("sitename")

	account, err := b.LookupAccount(u.ID, siteName)
	if err != nil {
		return fmt.Errorf("user %s has no account on %s", u.Username, siteName)
	}

	if err := b.UnlinkAccount(u.ID, account.Site.ID); err != nil {
		return err
	}

	return c.Redirect("/accounts")
}

func (b *Backend) ensureAccountNotRegistered(u *common.User, authUser *goth.User) error {
	stmt, err := b.Storage.Conn().Prepare(`
		SELECT accounts.username
		FROM accounts
		INNER JOIN sites ON accounts.site_id = sites.id
		WHERE sites.oauth_provider = ? AND accounts.username = ?
	`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(authUser.Provider, authUser.NickName)

	var username string
	if err = row.Scan(
		&username,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil
		} else {
			return err
		}
	}

	return fmt.Errorf("account is already registered")
}

func (b *Backend) RegisterAccount(u *common.User, authUser *goth.User) error {
	site, err := b.getSiteFromOAuthProvider(authUser.Provider)
	if err != nil {
		return err
	}

	stmt, err := b.Storage.Conn().Prepare(`
		INSERT INTO accounts (user_id, site_id, username, profile_url)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// TODO: make this not github-specific
	_, err = stmt.Exec(u.ID, site.ID, authUser.NickName, authUser.RawData["html_url"])
	if err != nil {
		return err
	}

	return nil
}

func (b *Backend) getSiteFromOAuthProvider(provider string) (*common.Site, error) {
	stmt, err := b.Storage.Conn().Prepare(`
		SELECT id
		FROM sites
		WHERE oauth_provider = ?
	`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(provider)

	var id int

	if err = row.Scan(&id); err != nil {
		return nil, fmt.Errorf("couldn't get OAuth provider: %w", err)
	}

	site, ok := b.Sites[id]
	if !ok {
		return nil, fmt.Errorf("site does not exist with found ID")
	}

	return site, nil
}

func (b *Backend) UnlinkAccount(userID int, siteID int) error {
	_, err := b.Storage.Conn().Exec(`
	DELETE FROM raw_points
	WHERE user_id = ?
	AND point_source_id IN (
		SELECT id
		FROM point_sources
		WHERE site_id = ?
	)`, userID, siteID)
	if err != nil {
		return err
	}

	_, err = b.Storage.Conn().Exec(`
	DELETE FROM accounts
	WHERE user_id = ? AND site_id = ?
	`, userID, siteID)

	return err
}
