package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"shorturl/internal/config"
	"shorturl/internal/handlers"
	"shorturl/internal/logger"
	"shorturl/internal/middleware"
	"shorturl/internal/storage/file"
	"shorturl/internal/storage/memory"
)

func main() {
	err := logger.Initialize("Info")
	if err != nil {
		panic(err)
	}

	config.ParseFlags()
	conf := config.GetConfig()

	// Create producer and consumer for storage.
	producer, err := file.NewProducer(conf.FileStoragePath)
	if err != nil {
		logger.Log.Fatal("Creating producer", zap.Error(err))
	}

	consumer, err := file.NewConsumer(conf.FileStoragePath)
	if err != nil {
		logger.Log.Fatal("Creating consumer", zap.Error(err))
	}

	// Close producer and consumer.
	defer func(producer *file.Producer) {
		err := producer.Close()
		if err != nil {
			logger.Log.Error("Closing producer", zap.Error(err))
		}
	}(producer)
	defer func(consumer *file.Consumer) {
		err := consumer.Close()
		if err != nil {
			logger.Log.Error("Closing consumer", zap.Error(err))

		}
	}(consumer)

	// Create storage and inject into handlers.
	storage := memory.NewStore(consumer, producer)
	handler := handlers.NewHandler(storage)

	logger.Log.Info("Reading stored data", zap.String("file", conf.FileStoragePath))
	err = storage.Load()
	if err != nil {
		logger.Log.Error("Load storage", zap.Error(err))
	}

	logger.Log.Info("Running server", zap.String("address", conf.RunAddr))

	err = http.ListenAndServe(conf.RunAddr, mainRouter(handler))
	if err != nil {
		panic(err)
	}
}

func mainRouter(h *handlers.Handler) chi.Router {
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
		r.Post("/", middleware.WithLogging(middleware.GzipMiddleware(h.GenerateURL))) // POST /
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", middleware.WithLogging(middleware.GzipMiddleware(h.GetURL))) // GET /EwHXdJfB
		})
		r.Route("/api/shorten", func(r chi.Router) {
			r.Post("/", middleware.WithLogging(middleware.GzipMiddleware(h.GenURLinJSON)))
		})
	})

	return r
}
