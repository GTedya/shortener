package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/GTedya/shortener/config"
	mock_repo "github.com/GTedya/shortener/internal/app/mocks"
	"github.com/GTedya/shortener/internal/app/tokenutils"
)

func TestDeleteUrls(t *testing.T) {
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
		body           interface{}
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Valid Request",
			body:           []string{"shortURL1", "shortURL2"},
			mockError:      nil,
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "Empty Body",
			body:           nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			body:           "invalid",
			mockError:      nil,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var requestBody io.Reader
			if test.body != nil {
				body, _ := json.Marshal(test.body)
				requestBody = bytes.NewReader(body)
			} else {
				requestBody = bytes.NewReader(nil)
			}

			req := httptest.NewRequest(http.MethodDelete, "http://localhost:8080/api/delete", requestBody)
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{
				Name:  tokenutils.UserIDCookieName,
				Value: "testUserID",
			})

			w := httptest.NewRecorder()

			if test.body != nil && test.body != "invalid" {
				mockRepo.EXPECT().DeleteUrls(gomock.Any(), gomock.Any()).Return(test.mockError).AnyTimes()
			}

			h.deleteUrls(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
		})
	}
}
