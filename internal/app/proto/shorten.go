package pb

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GTedya/shortener/internal/app/models"
	"github.com/GTedya/shortener/internal/app/repository"
)

// Shorten shortens the provided URL for the given user.
//
// If the URL in the request is empty, it returns an InvalidArgument error.
//
// If the user ID in the request is invalid or cannot be decoded and decrypted, it returns an InvalidArgument error.
//
// If the user ID is empty after decryption, a new user ID is generated.
//
// Calls the Shorten method on the service with the URL and user ID.
//
// If a duplicate URL is detected, it returns a response with the existing short URL.
//
// If an error occurs during the shortening process, it returns an Internal error.
//
// Parameters:
//   - ctx: The context for the request.
//   - r: The request containing the full URL and the user ID.
//
// Returns:
//   - A response containing the shortened URL and the user ID.
//   - An error if the URL is invalid, the user ID is invalid, or the shortening process fails.
func (s *Server) Shorten(ctx context.Context, r *ShortenRequest) (*ShorteningResponse, error) {
	if r.Url == "" {
		return nil, status.Error(codes.InvalidArgument, `full_url required`) //nolint:wrapcheck // it`s already wrapped
	}

	userID, err := s.decodeAndDecrypt(r.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, `invalid user_id`) //nolint:wrapcheck // it`s already wrapped
	}

	if userID == "" {
		userID = s.service.GenerateNewUserID()
	}

	shortURL, err := s.service.Shorten(ctx, r.Url, userID)
	if errors.Is(err, repository.ErrDuplicate) {
		// we cannot return "conflict" status with response, response becomes nil for client
		return s.newShorteningResponse(shortURL, ""), nil
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error()) //nolint:wrapcheck // it`s already wrapped
	}

	return s.newShorteningResponse(shortURL, userID), nil
}

// newShorteningResponse creates a new ShorteningResponse with the provided short URL and user ID.
//
// Parameters:
//   - shortURL: The short URL generated.
//   - userID: The user ID associated with the shortened URL.
//
// Returns:
//   - A ShorteningResponse containing the formatted short URL, user ID, and URL ID.
func (s *Server) newShorteningResponse(shortURL models.ShortURL, userID string) *ShorteningResponse {
	return &ShorteningResponse{
		ResultUrl: s.service.FormatShortURL(shortURL.ShortURL),
		UserId:    userID,
		UrlId:     shortURL.ShortURL,
	}
}
