package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"math/rand"
	"net/http"
	"net/url"
)

var URLMap map[string]string

func main() {
	URLMap = make(map[string]string)

	router := chi.NewRouter()

	router.Get("/{id}", getURLByID)
	router.Post("/", createURL)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func createURL(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType == "text/plain; charset=utf-8" {
		body, err := io.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(body) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id := GenerateTestURL(8)
		encodedID := url.PathEscape(id)
		URLMap[encodedID] = string(body)

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
	id := chi.URLParam(r, "id")

	link, ok := URLMap[id]

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	http.Redirect(w, r, link, http.StatusTemporaryRedirect)
}

var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateTestURL(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
