package pb

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GTedya/shortener/config"
	mock_service "github.com/GTedya/shortener/internal/app/mocks"
	"github.com/GTedya/shortener/internal/app/models"
)

func TestServer_ShortenBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_service.NewMockShortenerInterface(ctrl)
	conf := config.Config{SecretKey: "0123456789abcdef"}
	s := &Server{
		service: mockService,
		config:  conf,
	}

	t.Run("invalid user_id", func(t *testing.T) {
		req := &ShortenBatchRequest{UserId: "invalidUserId"}
		resp, err := s.ShortenBatch(context.Background(), req)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Equal(t, "invalid user_id", status.Convert(err).Message())
	})

	t.Run("successful batch shorten with new user_id", func(t *testing.T) {
		urls := []*ShortenBatchItemRequest{
			{OriginalUrl: "http://example1.com"},
			{OriginalUrl: "http://example2.com"},
		}
		req := &ShortenBatchRequest{
			Urls: urls,
		}

		batch := []models.ShortURL{
			{OriginalURL: "http://example1.com"},
			{OriginalURL: "http://example2.com"},
		}

		shortURLBatches := []models.ShortURL{
			{OriginalURL: "http://example1.com", ShortURL: "short1"},
			{OriginalURL: "http://example2.com", ShortURL: "short2"},
		}

		newUserID := "newUserId"
		mockService.EXPECT().GenerateNewUserID().Return(newUserID).Times(1)
		mockService.EXPECT().ShortenBatch(gomock.Any(), batch, newUserID).Return(shortURLBatches, nil).Times(1)
		mockService.EXPECT().FormatShortURL("short1").Return("http://localhost:8080/short1").Times(1)
		mockService.EXPECT().FormatShortURL("short2").Return("http://localhost:8080/short2").Times(1)

		resp, err := s.ShortenBatch(context.Background(), req)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Len(t, resp.Urls, len(urls))
		assert.Equal(t, "http://localhost:8080/short1", resp.Urls[0].ResultUrl)
		assert.Equal(t, "http://localhost:8080/short2", resp.Urls[1].ResultUrl)
	})
}
