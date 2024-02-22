package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/storage/sqlite3"
	"github.com/gofiber/template/html"
	"github.com/shareed2k/goth_fiber"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"

	"github.com/zac-garby/magicalinternetpoints/lib/backend"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
	"github.com/zac-garby/magicalinternetpoints/lib/integrations"
	"github.com/zac-garby/magicalinternetpoints/lib/integrations/providers"
	"github.com/zac-garby/magicalinternetpoints/lib/oauth-providers/reddit"
)

func main() {
	storage := sqlite3.New(sqlite3.Config{
		Database: "magicalinternetpoints.sqlite3",
	})

	backend, err := backend.New(storage)
	if err != nil {
		panic(err)
	}

	goth.UseProviders(
		github.New(
			os.Getenv("GITHUB_TOKEN"),
			os.Getenv("GITHUB_SECRET"),
			fmt.Sprintf("%s/auth/callback/github", os.Getenv("MIP_BASEURL")),
		),
		reddit.New(
			os.Getenv("REDDIT_TOKEN"),
			os.Getenv("REDDIT_SECRET"),
			fmt.Sprintf("%s/auth/callback/reddit", os.Getenv("MIP_BASEURL")),
		),
	)

	app := fiber.New(fiber.Config{
		Views:        html.New("./views", ".html"),
		ErrorHandler: errorHandler,
	})

	// static content
	app.Static("/static", "./static")

	// GET handlers
	app.Get("/", withUser(backend, func(user *common.User, c *fiber.Ctx) error {
		total, sources, err := backend.GetRawPoints(user.ID)
		if err != nil {
			return err
		}

		return c.Render("index", fiber.Map{
			"User":    user,
			"Sources": sources,
			"Total":   total,
		})
	}))

	app.Get("/metrics", monitor.New(monitor.Config{
		Title: "Magical Internet Points Metrics",
	}))

	app.Get("/rates",
		withUser(backend, func(user *common.User, c *fiber.Ctx) error {
			total, sources, err := backend.GetRawPoints(user.ID)
			if err != nil {
				return err
			}

			return c.Render("rates", fiber.Map{
				"Sites":   backend.Sites,
				"User":    user,
				"Sources": sources,
				"Total":   total,
			})
		}),
	)

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("login", fiber.Map{})
	})

	app.Get("/about", func(c *fiber.Ctx) error {
		return c.Render("about", fiber.Map{})
	})

	app.Get("/welcome", func(c *fiber.Ctx) error {
		return c.Render("welcome", fiber.Map{})
	})

	app.Get("/accounts",
		withUser(backend, func(user *common.User, c *fiber.Ctx) error {
			accounts, err := backend.GetAccounts(user.ID)
			if err != nil {
				return err
			}

			nonLinked, err := backend.GetNonLinkedSites(user.ID)
			if err != nil {
				return err
			}

			total, sources, err := backend.GetRawPoints(user.ID)
			if err != nil {
				return err
			}

			return c.Render("accounts", fiber.Map{
				"Accounts":  accounts,
				"NonLinked": nonLinked,
				"Sources":   sources,
				"Total":     total,
			})
		}),
	)

	app.Get("/auth/bio/:site_title",
		withUser(backend, func(user *common.User, c *fiber.Ctx) error {
			siteTitle := c.Params("site_title")

			site, ok := backend.GetSite(siteTitle)
			if !ok {
				return fmt.Errorf("site '%s' does not exist", siteTitle)
			}

			integration, ok := integrations.Integrations[siteTitle]
			if !ok {
				return fmt.Errorf("integration for site '%s' does not exist", siteTitle)
			}

			bioAuth, ok := integration.GetAuthProvider().(*providers.BioAuthProvider)
			if !ok {
				return fmt.Errorf("non-bio auth provider")
			}

			total, _, err := backend.GetRawPoints(user.ID)
			if err != nil {
				return err
			}

			username := c.Query("username")

			return c.Render("bio_auth", fiber.Map{
				"Site":       site,
				"VerifyText": fmt.Sprintf(providers.BIO_AUTH_STRING, user.Username),
				"Total":      total,
				"Username":   username, // maybe ""
				"ProfileURL": fmt.Sprintf(bioAuth.ProfileURL, username),
			})
		}),
	)

	// Registration
	app.Get("/auth/:site_title",
		withUser(backend, func(user *common.User, c *fiber.Ctx) error {
			return backend.BeginAuth(user, c)
		}),
	)

	// OAuth handlers
	app.Get("/auth/begin-oauth/:provider",
		withUser(backend, func(user *common.User, c *fiber.Ctx) error {
			return goth_fiber.BeginAuthHandler(c)
		}),
	)

	app.Get("/auth/callback/:provider",
		withUser(backend, backend.OAuthCallbackHandler),
	)

	// Bio auth handlers
	app.Get("/auth/bio/complete/:site_title/:username",
		withUser(backend, backend.BioAuthCompleteHandler),
	)

	// API handlers
	app.Post("/login", backend.AuthLoginHandler)
	app.Post("/register", backend.AuthRegisterHandler)
	app.Post("/logout", backend.AuthLogoutHandler)
	app.Post("/update/:sitename", withUser(backend, backend.UpdateHandler))
	app.Post("/unlink/:sitename", withUser(backend, backend.UnlinkHandler))

	app.Get("/logout", backend.AuthLogoutHandler)
	app.Get("/update/:sitename", withUser(backend, backend.UpdateHandler))
	app.Get("/unlink/:sitename", withUser(backend, backend.UnlinkHandler))
	app.Get("/badge/:username", backend.GetBadge)

	port := os.Getenv("MIP_PORT")
	if len(port) == 0 {
		port = "3000"
	}

	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}

func withUser(backend *backend.Backend, f func(user *common.User, c *fiber.Ctx) error) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user, err := backend.CurrentUser(c)
		if err != nil {
			return c.Redirect("/login")
		}

		return f(user, c)
	}
}

func errorHandler(c *fiber.Ctx, err error) error {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError

	// Retrieve the custom status code if it's a *fiber.Error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	err = c.Render("error", fiber.Map{
		"Code":    code,
		"Message": err.Error(),
	})

	if err != nil {
		// In case the SendFile fails
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return nil
}
