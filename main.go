package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
	"github.com/gofiber/template/html"

	"github.com/zac-garby/magicalinternetpoints/lib"
)

func main() {
	storage := sqlite3.New(sqlite3.Config{
		Database: "magicalinternetpoints.sqlite3",
	})

	sessions := session.New(session.Config{
		Storage: storage,
	})

	backend := lib.Backend{
		Storage:  storage,
		Sessions: sessions,
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
			return c.Render("index", fiber.Map{
				"User": user,
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
