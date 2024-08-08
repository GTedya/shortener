// Package pb provides the protocol buffer implementations for the service.
package pb

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/service"
	"github.com/GTedya/shortener/internal/app/tokenutils"
)

// Server represents the gRPC server for the URL shortener service.
type Server struct {
	UnimplementedShortenerServer
	server  *grpc.Server
	service service.ShortenerInterface
	config  config.Config
}

// NewGRPCServer creates a new instance of the gRPC server with the provided service and configuration.
//
// Parameters:
//   - service: The service implementing the ShortenerInterface.
//   - config: The configuration for the server.
//
// Returns:
//   - A pointer to the new Server instance.
//   - An error if the server could not be created.
func NewGRPCServer(
	service service.ShortenerInterface,
	config config.Config,
) (*Server, error) {
	s := grpc.NewServer()
	return &Server{
		server:  s,
		service: service,
		config:  config,
	}, nil
}

// Run starts the gRPC server on the specified address.
//
// Registers the ShortenerServer implementation and listens for incoming connections.
//
// Returns:
//   - An error if the server could not start or encountered an issue while running.
func (s *Server) Run() error {
	RegisterShortenerServer(s.server, s)

	listen, err := net.Listen("tcp", "localhost:3200")
	if err != nil {
		log.Fatal(err)
	}

	return s.server.Serve(listen) //nolint:wrapcheck // it returns error if the server could not start
}

// Shutdown gracefully stops the gRPC server.
//
// Returns:
//   - An error if the server could not be stopped gracefully.
func (s *Server) Shutdown() error {
	s.server.GracefulStop()
	return nil
}

// decodeAndDecrypt takes an encrypted and encoded hex string and returns the decoded and decrypted string.
//
// Parameters:
//   - userID: The encrypted and encoded user ID.
//
// Returns:
//   - The decoded and decrypted user ID as a string.
//   - An error if the decoding or decryption fails.
func (s *Server) decodeAndDecrypt(userID string) (string, error) {
	if userID == "" {
		return "", nil
	}
	decodedEncryptedUserID, err := hex.DecodeString(userID)
	if err != nil {
		return "", fmt.Errorf("decoding error: %w", err)
	}

	decryptedUserID, err := tokenutils.Decrypt(string(decodedEncryptedUserID), s.config.SecretKey)
	if err != nil {
		return "", fmt.Errorf("decryption error: %w", err)
	}

	return decryptedUserID, nil
}
