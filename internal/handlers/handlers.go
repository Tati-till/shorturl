package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"shorturl/internal/config"
	"shorturl/internal/hash"
	"shorturl/internal/logger"
	"shorturl/internal/models"
	"shorturl/internal/validation"
)

type Storage interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type Handler struct {
	storage Storage
}

func NewHandler(storage Storage) *Handler {
	return &Handler{storage: storage}
}

func (h *Handler) GenURLinJSON(res http.ResponseWriter, req *http.Request) {
	received, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, fmt.Sprintf("Can't read body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	var receivedReq models.Request
	err = json.Unmarshal(received, &receivedReq)
	if err != nil {
		http.Error(res, fmt.Sprintf("Can't unmarshal body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if !validation.IsCorrectURL(receivedReq.URL) {
		http.Error(res, "Invalid URL", http.StatusBadRequest)
		return
	}

	hashed, err := h.generator(receivedReq.URL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := models.Response{Result: hashed}
	resJSON, err := json.Marshal(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	_, err = res.Write(resJSON)
	if err != nil {
		logger.Log.Error("Failed to write response", zap.Error(err))
	}
}

func (h *Handler) generator(url string) (string, error) {
	hashed := hash.GetHashFromURL([]byte(url))
	err := h.storage.Set(hashed, url)
	if err != nil {
		return "", err
	}

	conf := config.GetConfig()
	return fmt.Sprintf("%s/%s", conf.ResAddr, hashed), nil
}

func (h *Handler) GenerateURL(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.Error(res, "Wrong request path", http.StatusBadRequest)
		return
	}

	receivedURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Can't read body", http.StatusBadRequest)
		return
	}

	strURL := string(receivedURL)
	if !validation.IsCorrectURL(strURL) {
		http.Error(res, "Invalid URL", http.StatusBadRequest)
		return
	}

	resURL, err := h.generator(strURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	_, err = res.Write([]byte(resURL))
	if err != nil {
		fmt.Println("Failed to write response:", err)
	}
}

func (h *Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	if id == "" {
		http.Error(res, "Wrong input URL", http.StatusBadRequest)
		return
	}

	storedURL, err := h.storage.Get(id)
	if err != nil {
		http.Error(res, "URL not found", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Location", storedURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
