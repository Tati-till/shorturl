package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"shorturl/internal/config"
	"shorturl/internal/logger"
	"shorturl/internal/models"
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

	if !isCorrectURL(receivedReq.URL) {
		http.Error(res, "Invalid URL", http.StatusBadRequest)
		return
	}

	hash, err := h.generator(receivedReq.URL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := models.Response{Result: hash}
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
	hash := getHashFromURL([]byte(url))
	err := h.storage.Set(hash, string(url))
	if err != nil {
		return "", err
	}

	conf := config.GetConfig()
	return fmt.Sprintf("%s/%s", conf.ResAddr, hash), nil
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
	if !isCorrectURL(strURL) {
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

func isCorrectURL(s string) bool {
	if s == "" {
		return false
	}
	_, err := url.Parse(s)
	return err == nil
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
