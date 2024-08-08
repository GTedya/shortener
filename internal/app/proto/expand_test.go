package pb

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/GTedya/shortener/config"
	mock_service "github.com/GTedya/shortener/internal/app/mocks"
	"github.com/GTedya/shortener/internal/app/models"
)

func TestServer_Expand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_service.NewMockShortenerInterface(ctrl)
	conf := config.Config{SecretKey: "0123456789abcdef"}
	s := &Server{
		service: mockService,
		config:  conf,
	}

	t.Run("missing url_id", func(t *testing.T) {
		req := &ExpandRequest{UrlId: ""}
		resp, err := s.Expand(context.Background(), req)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Equal(t, "url_id is required", status.Convert(err).Message())
	})

	t.Run("successful expansion", func(t *testing.T) {
		urlID := "validUrlID"
		originalURL := "http://example.com"
		shortURL := models.ShortURL{OriginalURL: originalURL, IsDeleted: false}

		mockService.EXPECT().Expand(gomock.Any(), urlID).Return(shortURL, nil).Times(1)

		req := &ExpandRequest{UrlId: urlID}
		resp, err := s.Expand(context.Background(), req)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, originalURL, resp.FullUrl)
	})

	t.Run("url_id not found", func(t *testing.T) {
		urlID := "invalidUrlID"
		shortURL := models.ShortURL{OriginalURL: "", IsDeleted: false}

		mockService.EXPECT().Expand(gomock.Any(), urlID).Return(shortURL, nil).Times(1)

		req := &ExpandRequest{UrlId: urlID}
		resp, err := s.Expand(context.Background(), req)
		assert.Nil(t, resp)
		assert.Equal(t, codes.NotFound, status.Code(err))
		assert.Equal(t, "url id is not found", status.Convert(err).Message())
	})

	t.Run("url is deleted", func(t *testing.T) {
		urlID := "deletedUrlID"
		originalURL := "http://example.com"
		shortURL := models.ShortURL{OriginalURL: originalURL, IsDeleted: true}

		mockService.EXPECT().Expand(gomock.Any(), urlID).Return(shortURL, nil).Times(1)

		req := &ExpandRequest{UrlId: urlID}
		resp, err := s.Expand(context.Background(), req)
		assert.Nil(t, resp)
		assert.Equal(t, codes.NotFound, status.Code(err))
		assert.Equal(t, "url is deleted", status.Convert(err).Message())
	})
}
