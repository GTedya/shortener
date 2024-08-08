package pb

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteUrls deletes the URLs specified in the request for the given user.
//
// If the user_id in the request is empty, it returns an InvalidArgument error.
//
// If the user_id cannot be decoded and decrypted, it returns an InvalidArgument error.
//
// It calls the DeleteUrls method on the service with the URL IDs and the decoded user ID.
//
// Parameters:
//   - ctx: The context for the request.
//   - r: The request containing the user ID and URL IDs to be deleted.
//
// Returns:
//   - An empty message if successful.
//   - An error if the user ID is invalid or cannot be processed.
func (s *Server) DeleteUrls(ctx context.Context, r *DeleteUrlsRequest) (*Empty, error) {
	if r.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, `user_id required`) //nolint:wrapcheck // it`s already wrapped
	}

	userID, err := s.decodeAndDecrypt(r.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, `invalid user_id`) //nolint:wrapcheck // it`s already wrapped
	}

	s.service.DeleteUrls(ctx, r.GetUrlIds(), userID)

	return &Empty{}, err
}
