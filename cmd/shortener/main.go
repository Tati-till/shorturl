package main

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"shorturl/internal/config"
	"shorturl/internal/logger"
	store "shorturl/internal/storage"
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
	err := logger.Initialize("Info")
	if err != nil {
		panic(err)
	}

	config.ParseFlags()
	conf := config.GetConfig()

	logger.Log.Info("Running server", zap.String("address", conf.RunAddr))

	err = http.ListenAndServe(conf.RunAddr, mainRouter())
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
		r.Post("/", logger.WithLogging(generateURL)) // POST /
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", logger.WithLogging(getURL)) // GET /EwHXdJfB
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
