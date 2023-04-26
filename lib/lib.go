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
