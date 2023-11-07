package main

import (
	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app"
	"github.com/GTedya/shortener/internal/helpers"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_createURL(t *testing.T) {
	conf := new(config.Config)
	conf.AnnounceConfig()

	data := helpers.URLData{
		URLMap: make(map[string]string),
	}

	type args struct {
		url         string
		method      string
		body        io.Reader
		contentType string
	}

	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "positive test #1",
			want: want{
				code:        201,
				contentType: "text/plain",
			},
			args: args{
				url:         "/",
				method:      http.MethodPost,
				body:        strings.NewReader(`https://yandex.ru`),
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.args.method, test.args.url, test.args.body)
			request.Header.Add("Content-Type", test.args.contentType)

			w := httptest.NewRecorder()
			app.CreateURL(w, request, conf, &data)

			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NotEmpty(t, resBody)

			require.NoError(t, err)

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func Test_getURLByID(t *testing.T) {
	data := helpers.URLData{
		URLMap: make(map[string]string),
	}
	data.URLMap["testID"] = "http://localhost:8080/testID"

	type args struct {
		url         string
		method      string
		body        io.Reader
		contentType string
	}

	type want struct {
		code        int
		location    string
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "pos test",
			want: want{
				code:        307,
				contentType: "text/plain",
				location:    data.URLMap["testID"],
			},
			args: args{
				url:         "http://localhost:8080/testID",
				method:      http.MethodGet,
				contentType: "text/plain",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/{id:[a-zA-Z0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
				app.GetURLByID(writer, request, data)
			})

			req := httptest.NewRequest(http.MethodGet, "/testID", nil)
			recorder := httptest.NewRecorder()

			r.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.location, res.Header.Get("location"))
		})
	}

}
