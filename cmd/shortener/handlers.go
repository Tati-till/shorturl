package main

import (
	"fmt"
	"io"
	"net/http"
)

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

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		body := fmt.Sprintf("%s%s/%s", host, port, hash)
		_, err = res.Write([]byte(body))
		if err != nil {
			http.Error(res, "Failed to write response", http.StatusInternalServerError)
			return
		}

	case http.MethodGet:
		input := req.URL.Path
		if len(input) > 1 && input[0] == '/' {
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
