package lib

import (
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
)

type Backend struct {
	Storage  *sqlite3.Storage
	Sessions *session.Store
}

type User struct {
	ID           int
	Email        string
	Username     string
	PasswordHash []byte
}

type Site struct {
	ID               int
	Title            string
	URL              string
	ScoreDescription string
}

type Account struct {
	Site       Site
	Username   string
	ProfileURL string
}
