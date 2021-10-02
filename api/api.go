package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

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

func RegisterAPI(r *mux.Router) error {
	api, err := MakeAPI(r)
	if err != nil {
		return err
	}

	api.Subrouter.Methods("POST").PathPrefix("/login").HandlerFunc(api.apiLogin)

	return nil
}

func (a *API) apiLogin(w http.ResponseWriter, r *http.Request) {
	var (
		email    = r.PostFormValue("email")
		password = r.PostFormValue("password")
		user     User
	)

	if err := a.DB.QueryRow("SELECT * FROM users WHERE email = ?", email).Scan(
		&user.ID, &user.Username, &user.Hash, &user.Salt, &user.Email); err != nil {
		if err == sql.ErrNoRows {
			errorResponse(w, "USER_DOES_NOT_EXIST", "No user exists with that email.", http.StatusBadRequest)
			return
		}

		errorResponse(w, "DATABASE_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		hash    = sha256.Sum256(append([]byte(password), user.Salt...))
		hexHash = fmt.Sprintf("%x", hash)
	)

	if hexHash == user.Hash {
		var (
			sessID = newSessionID()
			expiry = time.Now().Add(time.Hour * 24 * 7)
		)

		insert, err := a.DB.Query("INSERT INTO sessions VALUES (?, ?, ?)", sessID, user.ID, expiry)
		if err != nil {
			errorResponse(w, "DATABASE_ERROR", err.Error(), http.StatusInternalServerError)
			return
		}

		insert.Close()

		fmt.Fprint(w, sessID)
	} else {
		errorResponse(w, "INVALID_PASSWORD", "Invalid password.", http.StatusBadRequest)
		return
	}
}

func errorResponse(w http.ResponseWriter, code string, msg string, status int) {
	w.WriteHeader(status)

	resp := ErrorResponse{
		Code:    code,
		Message: msg,
	}

	j, _ := json.Marshal(&resp)

	w.Write(j)
}

func newSessionID() string {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]byte, 16)

	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}

	log.Println(string(s))

	return string(s)
}
