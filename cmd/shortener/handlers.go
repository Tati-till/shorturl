package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"shorturl/internal/config"
)

func generateURL(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.Error(res, "Wrong request path", http.StatusBadRequest)
		return
	}

	receivedURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Can't read body", http.StatusBadRequest)
		return
	}

	if !isCorrectURL(string(receivedURL)) {
		http.Error(res, "Invalid URL", http.StatusBadRequest)
	}

	hash := getHashFromURL(receivedURL)
	err = storageURLs.Set(hash, string(receivedURL))
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	conf := config.GetConfig()
	body := fmt.Sprintf("%s/%s", conf.ResAddr, hash)
	_, err = res.Write([]byte(body))
	if err != nil {
		fmt.Println("Failed to write response:", err)
	}
}

func getURL(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	if id == "" {
		http.Error(res, "Wrong input URL", http.StatusBadRequest)
		return
	}

	storedURL, err := storageURLs.Get(id)
	if err != nil {
		http.Error(res, "URL not found", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Location", storedURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func isCorrectURL(s string) bool {
	_, err := url.Parse(s)
	return err == nil
}
