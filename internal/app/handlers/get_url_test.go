package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GTedya/shortener/config"
	mock_repo "github.com/GTedya/shortener/internal/app/mocks"
	"github.com/GTedya/shortener/internal/app/models"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetURLByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repo.NewMockRepository(ctrl)
	h := &handler{
		repo: mockRepo,
		log:  zap.S(),
		conf: config.Config{URL: "http://example.com"},
	}

	tests := []struct {
		name            string
		id              string
		expectedStatus  int
		mockReturnURL   models.ShortURL
		mockReturnError error
	}{
		{
			name:           "Valid ID",
			id:             "validID",
			expectedStatus: http.StatusTemporaryRedirect,
			mockReturnURL: models.ShortURL{
				OriginalURL: "https://example.com",
			},
			mockReturnError: nil,
		},
		{
			name:            "URL Not Found",
			id:              "notFoundID",
			expectedStatus:  http.StatusBadRequest,
			mockReturnURL:   models.ShortURL{},
			mockReturnError: errors.New("URL not found"),
		},
		{
			name:           "URL Deleted",
			id:             "deletedID",
			expectedStatus: http.StatusGone,
			mockReturnURL: models.ShortURL{
				IsDeleted: true,
			},
			mockReturnError: errors.New("URL is deleted"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.EXPECT().GetByID(gomock.Any(), test.id).Return(test.mockReturnURL, test.mockReturnError)

			r := httptest.NewRequest(http.MethodGet, "http://example.com/"+test.id, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.id)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.getURLByID(w, r)

			assert.Equal(t, test.expectedStatus, w.Code)

			if test.expectedStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, test.mockReturnURL.OriginalURL, w.Header().Get("Location"))
			}
		})
	}
}

func BenchmarkGetURLByID(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mock_repo.NewMockRepository(ctrl)
	log := zap.NewExample().Sugar()
	conf := config.Config{URL: "http://localhost:8080"}

	h := &handler{
		repo: mockRepo,
		log:  log,
		conf: conf,
	}

	testID := "testID"
	testURL := models.ShortURL{
		OriginalURL: "http://localhost:8080/testID",
	}

	mockRepo.EXPECT().GetByID(gomock.Any(), testID).Return(testURL, nil).AnyTimes()

	r := chi.NewRouter()
	r.Get("/{id:[a-zA-Z0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		h.getURLByID(writer, request)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/"+testID, nil)
		recorder := httptest.NewRecorder()

		r.ServeHTTP(recorder, req)
	}
}
