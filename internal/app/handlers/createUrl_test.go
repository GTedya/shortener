package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/logger"
	"github.com/GTedya/shortener/internal/app/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateURL(t *testing.T) {
	conf := config.Config{Address: "localhost:8080", URL: "short"}

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

			store, err := storage.NewStore(conf, nil)
			if err != nil {
				t.Log(err)
			}

			h := &handler{log: log, conf: conf, store: store}
			h.createURL(w, request)

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

func BenchmarkCreateURL(b *testing.B) {
	conf := config.Config{Address: "localhost:8080", URL: "short"}
	log := logger.CreateLogger()
	store, err := storage.NewStore(conf, nil)
	if err != nil {
		b.Fatal(err)
	}

	h := &handler{log: log, conf: conf, store: store}

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`https://example.com`))
	request.Header.Add("Content-Type", "text/plain; charset=utf-8; application/json")
	w := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = io.NopCloser(strings.NewReader("example body"))
		h.createURL(w, request)

		if w.Code != http.StatusCreated {
			b.Fatalf("unexpected status code: got %d, want %d", w.Code, http.StatusCreated)
		}
	}
}

func TestJsonHandler(t *testing.T) {
	conf := config.Config{Address: "localhost:8080", URL: "short"}
	log := logger.CreateLogger()
	store, err := storage.NewStore(conf, nil)
	if err != nil {
		t.Log(err)
	}

	h := &handler{log: log, conf: conf, store: store}

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

			h.urlByJSON(w, request)
			assert.Equal(t, test.expectedStatus, w.Code)
		})
	}
}

func BenchmarkUrlByJSON(b *testing.B) {
	conf := config.Config{Address: "localhost:8080", URL: "short"}

	store, err := storage.NewStore(conf, nil)
	if err != nil {
		b.Fatal(err)
	}

	h := &handler{log: nil, conf: conf, store: store}
	reader := strings.NewReader(`{"url": "https://example.com"}`)
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten/", reader)
	request.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = io.NopCloser(reader)
		h.urlByJSON(w, request)
		if w.Code != http.StatusCreated {
			b.Fatalf("unexpected status code: got %d, want %d", w.Code, http.StatusCreated)
		}
	}
}

func TestBatch(t *testing.T) {
	conf := config.Config{
		Address: "localhost:8080",
		URL:     "short",
	}
	log := logger.CreateLogger()
	store, err := storage.NewStore(conf, nil)
	if err != nil {
		t.Fatal(err)
	}

	h := &handler{log: log, conf: conf, store: store}

	tests := []struct {
		name           string
		contentType    string
		body           string
		expectedStatus int
	}{
		{
			name:        "Valid Request",
			contentType: "application/json",
			body: `[{"original_url": "https://example.com", "correlation_id": "123456"},
{"original_url": "https://example2.com", "correlation_id": "1234567"}]`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid Content-Type",
			contentType:    "text/plain",
			body:           `[{"original_url": "https://example.com", "correlation_id": "654321"}]`,
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
			request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/batch", strings.NewReader(test.body))
			request.Header.Add("Content-Type", test.contentType)

			w := httptest.NewRecorder()
			h.batch(w, request)

			res := w.Result()
			defer func() {
				er := res.Body.Close()
				if er != nil {
					t.Fatal(er)
				}
			}()

			if res.StatusCode != test.expectedStatus {
				t.Errorf("Expected status code %d, got %d", test.expectedStatus, res.StatusCode)
			}

			if test.expectedStatus == http.StatusCreated {
				var response []storage.ResMultipleURL
				err = json.NewDecoder(res.Body).Decode(&response)
				if err != nil {
					t.Errorf("Error decoding response body: %v", err)
				}
				t.Log(len(response))
			}
		})
	}
}

func BenchmarkBatch(b *testing.B) {
	conf := config.Config{Address: "localhost:8080", URL: "short"}
	log := logger.CreateLogger()
	store, err := storage.NewStore(conf, nil)
	if err != nil {
		b.Fatal(err)
	}

	h := &handler{log: log, conf: conf, store: store}
	reader := strings.NewReader(`[{"original_url": "https://example.com", "correlation_id": "123456"},
{"original_url": "https://example2.com", "correlation_id": "1234567"}]`)

	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten/", reader)
	request.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = io.NopCloser(reader)
		h.batch(w, request)
	}
}
