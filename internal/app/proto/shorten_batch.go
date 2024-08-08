package pb

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GTedya/shortener/internal/app/models"
)

// ShortenBatch processes a batch of URLs and shortens them for the given user.
//
// If the user ID in the request is invalid or cannot be decoded and decrypted, it returns an InvalidArgument error.
//
// If the user ID is empty after decryption, a new user ID is generated.
//
// Each URL in the batch is validated, and an InvalidArgument error is returned if any URL is empty.
//
// Calls the ShortenBatch method on the service with the batch of URLs and user ID.
//
// If an error occurs during the batch shortening process, it returns an Internal error.
//
// Parameters:
//   - ctx: The context for the request.
//   - r: The request containing the batch of URLs and the user ID.
//
// Returns:
//   - A response containing the batch of shortened URLs.
//   - An error if any URL is invalid, the user ID is invalid, or the batch shortening process fails.
func (s *Server) ShortenBatch(ctx context.Context, r *ShortenBatchRequest) (*ShortenBatchResponse, error) {
	userID, err := s.decodeAndDecrypt(r.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, `invalid user_id`) //nolint:wrapcheck // it`s already wrapped
	}

	if userID == "" {
		userID = s.service.GenerateNewUserID()
	}

	batch := make([]models.ShortURL, len(r.GetUrls()))

	for i, url := range r.GetUrls() {
		if url.OriginalUrl == "" {
			return nil, status.Error(codes.InvalidArgument, `full_url required`) //nolint:wrapcheck // it`s already wrapped
		}
		batch[i] = models.ShortURL{
			OriginalURL: url.OriginalUrl,
		}
	}

	shortURLBatches, err := s.service.ShortenBatch(ctx, batch, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error()) //nolint:wrapcheck // it`s already wrapped
	}

	res := make([]*ShortenBatchItemResponse, len(shortURLBatches))
	for i, shortURLBatch := range shortURLBatches {
		res[i] = &ShortenBatchItemResponse{
			ResultUrl: s.service.FormatShortURL(shortURLBatch.ShortURL),
		}
	}

	return &ShortenBatchResponse{
		Urls: res,
	}, nil
}
