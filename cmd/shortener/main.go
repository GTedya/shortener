package main

import (
	"github.com/gorilla/mux"
	"io"
	"math/rand"
	"net/http"
)

var m map[string]string

func main() {
	m = make(map[string]string)
	router := mux.NewRouter()

	router.HandleFunc("/", createURL).Methods(http.MethodPost)

	router.HandleFunc("/{id}", getURLByID).Methods(http.MethodGet)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		panic(err)
	}
}

func createURL(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType == "text/plain" {
		url, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id := randStr(8)
		m[id] = string(url)
		w.Header().Set("Content-Type", "text/plain")

		http.Redirect(w, r, "http://localhost:8080/"+id, http.StatusCreated)
	}
}

func getURLByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	url, ok := m[id]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
