package main

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/storage/sqlite3"
	"github.com/gofiber/template/html"

	"github.com/zac-garby/magicalinternetpoints/lib/backend"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
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
		Views:        html.New("./views", ".html"),
		ErrorHandler: errorHandler,
	})

	// static content
	app.Static("/static", "./static")

	// GET handlers
	app.Get("/", withUser(backend, func(user *common.User, c *fiber.Ctx) error {
		sources, err := backend.GetRawPoints(user.ID)
		if err != nil {
			return err
		}

		return c.Render("index", fiber.Map{
			"User":    user,
			"Sources": sources,
		})
	}))

	app.Get("/metrics", monitor.New(monitor.Config{
		Title: "Magical Internet Points Metrics",
	}))

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("login", fiber.Map{})
	})

	app.Get("/about", func(c *fiber.Ctx) error {
		return c.Render("about", fiber.Map{})
	})

	app.Get("/welcome", func(c *fiber.Ctx) error {
		return c.Render("welcome", fiber.Map{})
	})

	app.Get("/logout", backend.AuthLogoutHandler)

	// POST handlers
	app.Post("/login", backend.AuthLoginHandler)
	app.Post("/register", backend.AuthRegisterHandler)
	app.Post("/logout", backend.AuthLogoutHandler)
	app.Post("/update/:sitename", backend.UpdateHandler)

	log.Fatal(app.Listen(":3000"))
}

func withUser(backend *backend.Backend, f func(user *common.User, c *fiber.Ctx) error) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user, err := backend.CurrentUser(c)
		if err != nil {
			return c.Redirect("/welcome")
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
