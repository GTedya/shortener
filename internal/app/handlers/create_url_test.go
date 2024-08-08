package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/logger"
	mock_repo "github.com/GTedya/shortener/internal/app/mocks"
)

func TestHandler_createURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repo.NewMockRepository(ctrl)
	h := &handler{
		repo: mockRepo,
		log:  zap.S(),
		conf: config.Config{
			URL: "http://localhost:8080",
		},
	}

	t.Run("empty request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rr := httptest.NewRecorder()

		h.createURL(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("successful creation", func(t *testing.T) {
		originalURL := "http://example.com"

		mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(originalURL)))
		rr := httptest.NewRecorder()

		h.createURL(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.NotEmpty(t, rr.Body.String())
	})
}

func BenchmarkCreateURL(b *testing.B) {
	conf := config.Config{Address: "localhost:8080", URL: "short"}
	log := logger.CreateLogger()
	ctrl := gomock.NewController(b)
	mockRepo := mock_repo.NewMockRepository(ctrl)

	h := &handler{log: log, conf: conf, repo: mockRepo}

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`https://example.com`))
	request.Header.Add("Content-Type", "text/plain; charset=utf-8; application/json")
	w := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = io.NopCloser(strings.NewReader("example body"))

		mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

		h.createURL(w, request)

		if w.Code != http.StatusCreated {
			b.Fatalf("unexpected status code: got %d, want %d", w.Code, http.StatusCreated)
		}
	}
}

func TestJsonHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repo.NewMockRepository(ctrl)
	h := &handler{
		repo: mockRepo,
		log:  zap.S(),
		conf: config.Config{
			URL: "http://localhost:8080",
		},
	}

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
			if test.contentType == "application/json" && test.body != "" {
				mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			}
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
	log := zap.S()

	ctrl := gomock.NewController(b)
	mockRepo := mock_repo.NewMockRepository(ctrl)

	h := &handler{log: log, conf: conf, repo: mockRepo}
	reader := strings.NewReader(`{"url": "https://example.com"}`)
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten/", reader)
	request.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = io.NopCloser(reader)

		mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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
	ctrl := gomock.NewController(t)
	mockRepo := mock_repo.NewMockRepository(ctrl)

	h := &handler{log: log, conf: conf, repo: mockRepo}

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

			if test.contentType == "application/json" && test.body != "" {
				mockRepo.EXPECT().SaveBatch(gomock.Any(), gomock.Any()).Return(nil)
			}

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
				var response []ResMultipleURL
				err := json.NewDecoder(res.Body).Decode(&response)
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
	ctrl := gomock.NewController(b)
	mockRepo := mock_repo.NewMockRepository(ctrl)

	h := &handler{log: log, conf: conf, repo: mockRepo}
	reader := strings.NewReader(`[{"original_url": "https://example.com", "correlation_id": "123456"},
{"original_url": "https://example2.com", "correlation_id": "1234567"}]`)

	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten/", reader)
	request.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = io.NopCloser(reader)
		mockRepo.EXPECT().SaveBatch(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		h.batch(w, request)
	}
}
