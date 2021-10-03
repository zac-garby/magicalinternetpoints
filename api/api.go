package api

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

type API struct {
	Subrouter *mux.Router
	DB        *sql.DB
}

type User struct {
	ID                          int
	Username, Hash, Salt, Email string
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func MakeAPI(r *mux.Router) (*API, error) {
	a := &API{
		Subrouter: r.PathPrefix("/api/").Subrouter(),
	}

	var (
		username = os.Getenv("MIP_DB_USERNAME")
		password = os.Getenv("MIP_DB_PASSWORD")
		ip       = os.Getenv("MIP_DB_IP")
		port     = os.Getenv("MIP_DB_PORT")
		dbName   = os.Getenv("MIP_DB_DATABASE")
	)

	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, ip, port, dbName),
	)

	if err != nil {
		return nil, err
	}

	a.DB = db

	return a, nil
}
