package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"shorturl/internal/config"
)

func generateURL(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.Error(res, "Wrong request path", http.StatusBadRequest)
		return
	}

	url, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Can't read body", http.StatusBadRequest)
		return
	}

	hash := getHashFromURL(url)
	err = storageURLs.Set(hash, string(url))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)

	conf := config.GetConfig()
	body := fmt.Sprintf("%s/%s", conf.ResAddr, hash)
	_, err = res.Write([]byte(body))
	if err != nil {
		http.Error(res, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func getURL(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	if id == "" {
		http.Error(res, "Wrong input URL", http.StatusBadRequest)
		return
	}

	url, err := storageURLs.Get(id)
	if err != nil {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
