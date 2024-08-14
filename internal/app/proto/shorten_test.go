package pb

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GTedya/shortener/config"
	mock_service "github.com/GTedya/shortener/internal/app/mocks"
	"github.com/GTedya/shortener/internal/app/models"
)

func TestServer_Shorten(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_service.NewMockShortenerInterface(ctrl)
	conf := config.Config{SecretKey: "0123456789abcdef0123456789abcdef"} // 32-байтовый ключ для AES-256
	s := &Server{
		service: mockService,
		config:  conf,
	}

	t.Run("missing full_url", func(t *testing.T) {
		req := &ShortenRequest{UserId: "validUserId"}
		resp, err := s.Shorten(context.Background(), req)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Equal(t, "full_url required", status.Convert(err).Message())
	})

	t.Run("invalid user_id", func(t *testing.T) {
		req := &ShortenRequest{Url: "http://example.com", UserId: "invalidUserId"}
		resp, err := s.Shorten(context.Background(), req)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Equal(t, "invalid user_id", status.Convert(err).Message())
	})

	t.Run("successful shorten with new user_id", func(t *testing.T) {
		originalURL := "http://example.com"
		shortID := uuid.NewString()
		newUserID := "newUserID"

		mockService.EXPECT().GenerateNewUserID().Return(newUserID).Times(1)
		mockService.EXPECT().Shorten(gomock.Any(), originalURL, newUserID).Return(models.ShortURL{
			OriginalURL: originalURL,
			ShortURL:    shortID,
		}, nil).Times(1)
		mockService.EXPECT().FormatShortURL(shortID).Return("http://localhost:8080/" + shortID).Times(1)

		req := &ShortenRequest{Url: originalURL}
		resp, err := s.Shorten(context.Background(), req)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, "http://localhost:8080/"+shortID, resp.ResultUrl)
		assert.Equal(t, newUserID, resp.UserId)
		assert.Equal(t, shortID, resp.UrlId)
	})
}
