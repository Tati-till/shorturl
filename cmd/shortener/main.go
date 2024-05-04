package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	store "shorturl/internal/storage"
)

const (
	host = "http://localhost"
	port = ":8080"
)

var storageURLs store.Store

func mainHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
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

		res.WriteHeader(http.StatusCreated)
		body := fmt.Sprintf("%s%s/%s", host, port, hash)
		_, err = res.Write([]byte(body))
		if err != nil {
			http.Error(res, "Failed to write response", http.StatusInternalServerError)
			return
		}

	case http.MethodGet:
		input := req.URL.Path
		if len(input) > 0 && input[0] == '/' {
			input = input[1:]
		} else {
			http.Error(res, "Wrong input URL", http.StatusBadRequest)
			return
		}

		url, err := storageURLs.Get(input)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", url)
		res.WriteHeader(http.StatusTemporaryRedirect)

	default:
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
}

func main() {
	var err error

	storageURLs, err = store.NewStore()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainHandler)

	err = http.ListenAndServe(port, mux)
	if err != nil {
		panic(err)
	}
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
