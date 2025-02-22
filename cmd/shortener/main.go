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
	"shorturl/internal/storage/file"
)

var (
	storageURLs store.Store
	producer    *file.Producer
	consumer    *file.Consumer
)

func main() {
	err := logger.Initialize("Info")
	if err != nil {
		panic(err)
	}

	config.ParseFlags()
	conf := config.GetConfig()

	logger.Log.Info("Reading stored data", zap.String("file", conf.FileStoragePath))
	initStorage(conf)

	// Close producer and consumer.
	defer func(Producer *file.Producer) {
		err := Producer.Close()
		if err != nil {
			logger.Log.Error("Closing producer", zap.Error(err))
		}
	}(producer)
	defer func(Consumer *file.Consumer) {
		err := Consumer.Close()
		if err != nil {
			logger.Log.Error("Closing consumer", zap.Error(err))

		}
	}(consumer)

	logger.Log.Info("Running server", zap.String("address", conf.RunAddr))

	err = http.ListenAndServe(conf.RunAddr, mainRouter())
	if err != nil {
		panic(err)
	}
}

func initStorage(conf *config.Config) {
	var err error

	producer, err = file.NewProducer(conf.FileStoragePath)
	if err != nil {
		logger.Log.Fatal("Creating producer", zap.Error(err))
	}

	consumer, err = file.NewConsumer(conf.FileStoragePath)
	if err != nil {
		logger.Log.Fatal("Creating consumer", zap.Error(err))
	}

	storageURLs, err = store.NewStore(consumer, producer)
	if err != nil {
		logger.Log.Fatal("New storage", zap.Error(err))
	}

	err = storageURLs.Load()
	if err != nil {
		logger.Log.Error("Load storage", zap.Error(err))
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
		r.Post("/", logger.WithLogging(gzipMiddleware(generateURL))) // POST /
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", logger.WithLogging(gzipMiddleware(getURL))) // GET /EwHXdJfB
		})
		r.Route("/api/shorten", func(r chi.Router) {
			r.Post("/", logger.WithLogging(gzipMiddleware(genURLinJSON)))
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
