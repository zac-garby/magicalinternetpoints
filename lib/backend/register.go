package backend

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gofiber/fiber/v2"
	"github.com/markbates/goth"
	"github.com/shareed2k/goth_fiber"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
	"github.com/zac-garby/magicalinternetpoints/lib/integrations"
	"github.com/zac-garby/magicalinternetpoints/lib/integrations/providers"
)

func (b *Backend) BeginAuth(u *common.User, c *fiber.Ctx) error {
	var (
		siteTitle       = c.Params("site_title")
		integration, ok = integrations.Integrations[siteTitle]
	)

	if !ok {
		return fmt.Errorf("authentication not implemented for site '%s'", siteTitle)
	}

	provider := integration.GetAuthProvider()
	return provider.BeginAuthentication(u, c)
}

func (b *Backend) OAuthCallbackHandler(u *common.User, c *fiber.Ctx) error {
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

	site, ok := b.getSiteFromOAuthProvider(authUser.Provider)
	if !ok {
		return fmt.Errorf("no provider for oauth %s", authUser.Provider)
	}

	if err := b.RegisterAccount(u, authUser.NickName, site); err != nil {
		return fmt.Errorf("could not register account: %w", err)
	}

	return c.Redirect("/accounts")
}

func (b *Backend) BioAuthCompleteHandler(u *common.User, c *fiber.Ctx) error {
	var (
		col             = colly.NewCollector()
		requiredText    = fmt.Sprintf(providers.BIO_AUTH_STRING, u.Username)
		siteTitle       = c.Params("site_title")
		username        = c.Params("username")
		authenticated   = false
		integration, ok = integrations.Integrations[siteTitle]
	)

	if !ok {
		return fmt.Errorf("authentication not implemented for site '%s'", siteTitle)
	}

	if username == "" {
		return fmt.Errorf("no username provided for site %s", siteTitle)
	}

	bioProvider, ok := integration.GetAuthProvider().(*providers.BioAuthProvider)
	if !ok {
		return fmt.Errorf("bio provider not found for site %s", siteTitle)
	}

	profileURL := fmt.Sprintf(bioProvider.ProfileURL, username)

	col.OnHTML("p#portfolio-user-bio", func(e *colly.HTMLElement) {
		content := e.ChildText("*")
		authenticated = strings.Contains(content, requiredText)
	})

	col.Visit(profileURL)
	col.Wait()

	if authenticated {
		site, ok := b.GetSite(siteTitle)
		if !ok {
			return fmt.Errorf("could not authenticate: site %s doesn't exist?", siteTitle)
		}

		b.RegisterAccount(u, username, site)

		return c.Redirect("/accounts")
	} else {
		return fmt.Errorf("authentication failed: please make sure the text '%s' is in your bio somewhere at %s", requiredText, profileURL)
	}
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

func (b *Backend) RegisterAccount(u *common.User, username string, site *common.Site) error {
	stmt, err := b.Storage.Conn().Prepare(`
		INSERT INTO accounts (user_id, site_id, username, profile_url)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	profileURL := integrations.Integrations[site.Title].GetProfileURL(username)

	_, err = stmt.Exec(u.ID, site.ID, username, profileURL)
	if err != nil {
		return err
	}

	// finally, update the points to get an initial score
	return b.UpdatePoints(u.ID, site.Title)
}

func (b *Backend) ensureAccountNotRegistered(u *common.User, authUser *goth.User) error {
	site, ok := b.getSiteFromOAuthProvider(authUser.Provider)
	if !ok {
		return fmt.Errorf("oauth provider not registered for %s", authUser.Provider)
	}

	stmt, err := b.Storage.Conn().Prepare(`
		SELECT accounts.username
		FROM accounts
		INNER JOIN sites ON accounts.site_id = sites.id
		WHERE sites.id = ? AND accounts.username = ?
	`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(site.ID, authUser.NickName)

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

func (b *Backend) getSiteFromOAuthProvider(provider string) (*common.Site, bool) {
	// TODO: replace this with a lookup in b.Sites?

	for _, site := range b.Sites {
		integration, ok := integrations.Integrations[site.Title]
		if !ok {
			continue
		}

		oauth, ok := integration.GetAuthProvider().(*providers.OAuthProvider)
		if !ok {
			continue
		}

		if oauth.ProviderName == provider {
			return site, true
		}
	}

	return nil, false
}

func (b *Backend) GetSite(title string) (*common.Site, bool) {
	for _, site := range b.Sites {
		if site.Title == title {
			return site, true
		}
	}

	return nil, false
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
