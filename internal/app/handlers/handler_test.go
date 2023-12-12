package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/helpers"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_createURL(t *testing.T) {
	conf := config.Config{Address: "localhost:8080", URL: "short"}
	data := make(map[string]string)

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
				contentType: "text/plain; application/json",
			},
			args: args{
				url:         "/",
				method:      http.MethodPost,
				body:        strings.NewReader(`https://yandex.ru`),
				contentType: "text/plain; charset=utf-8; application/json",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.args.method, test.args.url, test.args.body)
			request.Header.Add("Content-Type", test.args.contentType)

			w := httptest.NewRecorder()
			log := logger.CreateLogger()

			h := &handler{log: log, conf: conf, store: helpers.NewStore(conf, data)}
			h.CreateURL(w, request)

			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)

			defer func() {
				err := res.Body.Close()
				if err != nil {
					t.Log(fmt.Errorf("response body closing error: %w", err))
				}
			}()

			resBody, err := io.ReadAll(res.Body)
			require.NotEmpty(t, resBody)

			require.NoError(t, err)

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func Test_getURLByID(t *testing.T) {
	data := make(map[string]string)
	data["testID"] = "http://localhost:8080/testID"

	type args struct {
		url         string
		method      string
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
				contentType: "text/plain; application/json",
				location:    data["testID"],
			},
			args: args{
				url:         "http://localhost:8080/testID",
				method:      http.MethodGet,
				contentType: "text/plain; application/json",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := chi.NewRouter()
			conf := config.Config{Address: "localhost:8080", URL: "short"}

			h := &handler{store: helpers.NewStore(conf, data)}
			r.Get("/{id:[a-zA-Z0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
				h.GetURLByID(writer, request)
			})

			req := httptest.NewRequest(http.MethodGet, "/testID", nil)
			recorder := httptest.NewRecorder()

			r.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer func() {
				err := res.Body.Close()
				if err != nil {
					t.Log(fmt.Errorf("response body closing error: %w", err))
				}
			}()

			assert.Equal(t, test.want.code, res.StatusCode)

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.location, res.Header.Get("location"))
		})
	}
}

func TestJsonHandler(t *testing.T) {
	conf := config.Config{Address: "localhost:8080", URL: "short"}
	data := make(map[string]string)
	log := logger.CreateLogger()
	h := &handler{log: log, conf: conf, store: helpers.NewStore(conf, data)}

	// Test cases
	tests := []struct {
		name           string
		contentType    string
		body           string
		expectedStatus int
	}{
		{
			name:           "Valid JSON",
			contentType:    "application/json",
			body:           `{"url": "https://example.com"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid Content-Type",
			contentType:    "text/plain",
			body:           `{"url": "https://example.com"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty Body",
			contentType:    "application/json",
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten/", strings.NewReader(test.body))
			request.Header.Add("Content-Type", test.contentType)

			w := httptest.NewRecorder()

			h.URLByJSON(w, request)
			assert.Equal(t, test.expectedStatus, w.Code)
		})
	}
}
