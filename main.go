package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/sqlite3"
	"github.com/gofiber/template/html"

	"github.com/zac-garby/magicalinternetpoints/lib/backend"
)

func main() {
	storage := sqlite3.New(sqlite3.Config{
		Database: "magicalinternetpoints.sqlite3",
	})

	backend, err := backend.New(storage)
	if err != nil {
		panic(err)
	}

	app := fiber.New(fiber.Config{
		Views: html.New("./views", ".html"),
	})

	// static content
	app.Static("/static", "./static")

	// GET handlers
	app.Get("/", func(c *fiber.Ctx) error {
		user, err := backend.CurrentUser(c)

		if err != nil {
			return c.Render("welcome", fiber.Map{})
		} else {
			// accounts, err := backend.LookupAccounts(user.ID)
			// if err != nil {
			// 	panic(err)
			// }

			sources, err := backend.GetRawPoints(user.ID)
			if err != nil {
				return err
			}

			return c.Render("index", fiber.Map{
				"User":    user,
				"Sources": sources,
			})
		}
	})

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("login", fiber.Map{})
	})

	app.Get("/logout", backend.AuthLogoutHandler)

	// POST handlers
	app.Post("/login", backend.AuthLoginHandler)
	app.Post("/register", backend.AuthRegisterHandler)
	app.Post("/logout", backend.AuthLogoutHandler)

	log.Fatal(app.Listen(":3000"))
}
