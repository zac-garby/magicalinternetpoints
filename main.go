package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/zac-garby/magicalinternetpoints/api"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	rand.Seed(time.Now().Unix())

	serv, m := makeHTTPSServer()

	log.Println("listening on :80")
	go http.ListenAndServe(":80", m.HTTPHandler(nil))

	log.Println("listening on :443")
	if err := serv.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("ListenAndServeTLS() failed with %s\n", err)
	}
}

func makeHTTPSServer() (*http.Server, *autocert.Manager) {
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("magicalinternetpoints.com", "www.magicalinternetpoints.com"),
		Cache:      autocert.DirCache("secret"),
	}

	serv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      makeRouter(),
		Addr:         ":443",
		TLSConfig:    m.TLSConfig(),
	}

	return serv, m
}

func makeRouter() *mux.Router {
	r := mux.NewRouter()

	if err := api.RegisterAPI(r); err != nil {
		log.Fatalf("could not create API. %s", err)
	}

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	r.Path("/login").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/login.html")
	})

	r.Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	return r
}
