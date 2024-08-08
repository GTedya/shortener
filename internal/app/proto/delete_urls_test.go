package pb

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GTedya/shortener/config"
	mock_service "github.com/GTedya/shortener/internal/app/mocks"
	"github.com/GTedya/shortener/internal/app/tokenutils"
)

func TestServer_DeleteUrls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_service.NewMockShortenerInterface(ctrl)
	// Correcting the SecretKey length to 16 bytes (128-bit key)
	conf := config.Config{SecretKey: "0123456789abcdef"}
	s := &Server{
		service: mockService,
		config:  conf,
	}

	t.Run("missing user_id", func(t *testing.T) {
		req := &DeleteUrlsRequest{
			UserId: "",
			UrlIds: []string{"url1", "url2"},
		}
		resp, err := s.DeleteUrls(context.Background(), req)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Equal(t, "user_id required", status.Convert(err).Message())
	})

	t.Run("invalid user_id", func(t *testing.T) {
		invalidUserID := "invalidUserId"
		req := &DeleteUrlsRequest{
			UserId: invalidUserID,
			UrlIds: []string{"url1", "url2"},
		}

		resp, err := s.DeleteUrls(context.Background(), req)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Equal(t, "invalid user_id", status.Convert(err).Message())
	})

	t.Run("successful deletion", func(t *testing.T) {
		userID := "validUserId"
		encryptedUserID, err := tokenutils.Encrypt(userID, s.config.SecretKey)
		if err != nil {
			t.Fatalf("Failed to encrypt user ID: %v", err)
		}
		encodedUserID := hex.EncodeToString([]byte(encryptedUserID))

		req := &DeleteUrlsRequest{
			UserId: encodedUserID,
			UrlIds: []string{"url1", "url2"},
		}
		mockService.EXPECT().DeleteUrls(gomock.Any(), req.GetUrlIds(), userID).Return().Times(1)

		resp, err := s.DeleteUrls(context.Background(), req)
		assert.NotNil(t, resp)
		assert.NoError(t, err)
	})
}
