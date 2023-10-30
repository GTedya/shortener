package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"math/rand"
	"net/http"
	"net/url"
)

var urlMap map[string]string

func main() {
	urlMap = make(map[string]string)

	router := mux.NewRouter()

	router.HandleFunc("/{id}", getURLByID).Methods(http.MethodGet)
	router.HandleFunc("/", createURL).Methods(http.MethodPost)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func createURL(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType == "text/plain" {
		body, err := io.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(body) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id := generateTestURL(8)
		encodedID := url.PathEscape(id)
		urlMap[encodedID] = string(body)

		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)

		_, err = w.Write([]byte("http://localhost:8080/" + encodedID))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func getURLByID(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType == "text/plain" {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		link, ok := urlMap[id]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Add("Content-Type", "text/plain")
		http.Redirect(w, r, link, http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateTestURL(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
