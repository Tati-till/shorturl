package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"shorturl/internal/models"

	"github.com/go-chi/chi/v5"
	"shorturl/internal/config"
)

func genURLinJSON(res http.ResponseWriter, req *http.Request) {
	received, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Can't read body", http.StatusBadRequest)
		return
	}

	var receivedReq models.Request
	err = json.Unmarshal(received, &receivedReq)
	if err != nil {
		http.Error(res, "Can't read body", http.StatusBadRequest)
		return
	}

	if receivedReq.URL == "" {
		http.Error(res, "Invalid URL", http.StatusBadRequest)
		return
	}

	hash, err := generator(receivedReq.URL)
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
		fmt.Println("Failed to write response:", err)
	}
}

func generator(url string) (string, error) {
	hash := getHashFromURL([]byte(url))
	err := storageURLs.Set(hash, string(url))
	if err != nil {
		return "", err
	}

	conf := config.GetConfig()
	return fmt.Sprintf("%s/%s", conf.ResAddr, hash), nil
}

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

	strURL := string(receivedURL)
	if !isCorrectURL(strURL) {
		http.Error(res, "Invalid URL", http.StatusBadRequest)
		return
	}

	resURL, err := generator(strURL)
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
