package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GTedya/shortener/config"
	mock_repo "github.com/GTedya/shortener/internal/app/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetPing(t *testing.T) {
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
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Ping Success",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.EXPECT().Check(gomock.Any()).Return(test.mockError).AnyTimes()

			req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/ping", nil)
			w := httptest.NewRecorder()

			h.getPing(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
		})
	}
}
