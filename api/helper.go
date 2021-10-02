package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"time"
)

var (
	ErrNoSession      = errors.New("no session exists")
	ErrSessionExpired = errors.New("session expired")
)

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

	return string(s)
}

func (a *API) getSession(r *http.Request) (*User, error) {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		return nil, ErrNoSession
	}

	var (
		sessionID = sessionCookie.Value
		user      User
		expiryStr string
	)

	err = a.DB.QueryRow("SELECT id, username, password_hash, password_salt, email, expiry FROM users INNER JOIN sessions WHERE sess_id = ?",
		sessionID).Scan(&user.ID, &user.Username, &user.Hash, &user.Salt, &user.Email, &expiryStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoSession
		}

		return nil, err
	}

	expiry, err := time.Parse("2006-01-02 15:04:05", expiryStr)
	if err != nil {
		panic("this shouldn't happen. " + err.Error())
	}

	if time.Now().Before(expiry) {
		return &user, nil
	} else {
		// should probably delete the session from the database here
		return nil, ErrSessionExpired
	}
}
