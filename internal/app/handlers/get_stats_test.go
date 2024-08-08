package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/GTedya/shortener/config"
	mock_repo "github.com/GTedya/shortener/internal/app/mocks"
	"github.com/GTedya/shortener/internal/app/models"
)

func TestGetStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repo.NewMockRepository(ctrl)
	log := zap.NewExample().Sugar()
	h := &handler{
		repo: mockRepo,
		log:  log,
		conf: config.Config{URL: "http://localhost:8080"},
	}

	tests := []struct {
		name           string
		mockUserCount  int
		mockUrlsCount  int
		mockError      error
		expectedStatus int
		expectedBody   models.Stats
	}{
		{
			name:           "Valid Stats",
			mockUserCount:  10,
			mockUrlsCount:  100,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   models.Stats{UsersCount: 10, UrlsCount: 100},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.EXPECT().GetUsersAndUrlsCount(gomock.Any()).
				Return(test.mockUserCount, test.mockUrlsCount, test.mockError).AnyTimes()

			req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/stats", nil)
			w := httptest.NewRecorder()

			h.getStats(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)

			if test.expectedStatus == http.StatusOK {
				var result models.Stats
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, test.expectedBody, result)
			}
		})
	}
}
