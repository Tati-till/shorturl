package main

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"github.com/go-chi/chi/v5"
	store "shorturl/internal/storage"
)

const (
	host = "http://localhost"
	port = ":8080"
)

func init() {
	var err error
	storageURLs, err = store.NewStore()
	if err != nil {
		panic(err)
	}
}

var storageURLs store.Store

func main() {
	err := http.ListenAndServe(port, mainRouter())
	if err != nil {
		panic(err)
	}
}

func mainRouter() chi.Router {
	r := chi.NewRouter()

	// Custom handler for unsupported routes
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	})

	// Custom handler for unsupported methods
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	})

	r.Route("/", func(r chi.Router) {
		r.Post("/", generateURL) // POST /
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", getURL) // GET /EwHXdJfB
		})
	})

	return r
}

func getHashFromURL(url []byte) string {
	hasher := sha256.New()
	hasher.Write(url)
	hashBytes := hasher.Sum(nil)

	// Encode the first 6 bytes of the hash to base64
	// 6 bytes are chosen to ensure that the base64 encoded string is at least 8 characters long
	shortHash := base64.RawURLEncoding.EncodeToString(hashBytes[:6])
	return shortHash
}
