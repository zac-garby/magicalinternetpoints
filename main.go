package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/zac-garby/magicalinternetpoints/api"
)

func main() {
	rand.Seed(time.Now().Unix())

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

	log.Println("listening on http://localhost:8000")
	http.ListenAndServe(":8000", r)
}
