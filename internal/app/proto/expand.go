package pb

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Expand expands the short URL specified in the request to its original form.
//
// If the url_id in the request is empty, it returns an InvalidArgument error.
//
// If the service encounters an error while expanding the URL, it returns an Internal error.
//
// If the expanded URL is not found or is deleted, it returns a NotFound error.
//
// Parameters:
//   - ctx: The context for the request.
//   - r: The request containing the URL ID to be expanded.
//
// Returns:
//   - The full URL in the response if successful.
//   - An error if the URL ID is invalid, the URL is not found, or the URL is deleted.
func (s *Server) Expand(ctx context.Context, r *ExpandRequest) (*ExpandResponse, error) {
	urlID := r.GetUrlId()
	if urlID == "" {
		return nil, status.Error(codes.InvalidArgument, "url_id is required") //nolint:wrapcheck // it`s already wrapped
	}
	shortURL, err := s.service.Expand(ctx, urlID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error()) //nolint:wrapcheck // it`s already wrapped
	}

	if shortURL.OriginalURL == "" {
		return nil, status.Error(codes.NotFound, "url id is not found") //nolint:wrapcheck // it`s already wrapped
	}

	if shortURL.IsDeleted {
		return nil, status.Error(codes.NotFound, "url is deleted") //nolint:wrapcheck // it`s already wrapped
	}

	return &ExpandResponse{
		FullUrl: shortURL.OriginalURL,
	}, nil
}
