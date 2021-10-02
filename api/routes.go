package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func RegisterAPI(r *mux.Router) error {
	api, err := MakeAPI(r)
	if err != nil {
		return err
	}

	api.Subrouter.Methods("POST").PathPrefix("/login").HandlerFunc(api.apiLogin)
	api.Subrouter.Methods("POST").PathPrefix("/get-user").HandlerFunc(api.apiGetUser)

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

		fmt.Fprintf(w, `{"session_id": "%s", "expiry": "%s"}`, sessID, expiry.UTC())
	} else {
		errorResponse(w, "INVALID_PASSWORD", "Invalid password.", http.StatusBadRequest)
		return
	}
}

func (a *API) apiGetUser(w http.ResponseWriter, r *http.Request) {
	user, err := a.getSession(r)
	if err != nil {
		if err == ErrNoSession {
			errorResponse(w, "NO_SESSION", "Your session ID is invalid.", http.StatusBadRequest)
		} else if err == ErrSessionExpired {
			errorResponse(w, "SESSION_EXPIRED", "Your session has expired.", http.StatusBadRequest)
		} else {
			errorResponse(w, "DATABASE_ERROR", err.Error(), http.StatusInternalServerError)
		}

		return
	}

	var response struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	response.Username = user.Username
	response.Email = user.Email

	j, _ := json.Marshal(&response)
	w.Write(j)
}
