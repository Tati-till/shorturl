package main

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"

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
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainHandler)

	err := http.ListenAndServe(port, mux)
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
