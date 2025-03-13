package blobrpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/dnldd/blobrpc/rpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// Executor processes a request.
type Executor func(ctx context.Context, req *rpc.Request) (*rpc.Response, error)

// RequestHandler defines the requirements of a blobrpc request handler.
type RequestHandler interface {
	// ID generates the executor id from the provided route and version.
	ID(route string, version uint32) string
	// FetchExecutor fetches the executor for the provided id.
	FetchExecutor(id string) (Executor, bool)
}

// ServerConfig represents the configuration of the blobrpc server.
type ServerConfig struct {
	// The listening port of the server.
	Port uint64
	// The grpc server configuration options.
	Options []grpc.ServerOption
	// The request handler.
	Handler RequestHandler
	// The service logger.
	Logger *zerolog.Logger
}

// Server represents a blobrpc server.
type Server struct {
	cfg *ServerConfig
	gs  *grpc.Server

	rpc.BlobServer
}

// NewServer initializes a new blobrpc server.
func NewServer(cfg *ServerConfig) *Server {
	return &Server{
		cfg: cfg,
		gs:  grpc.NewServer(cfg.Options...),
	}
}

// Send processes the provided request.
func (s *Server) Send(ctx context.Context, req *rpc.Request) (*rpc.Response, error) {
	id := s.cfg.Handler.ID(req.Route, req.Version)
	f, ok := s.cfg.Handler.FetchExecutor(id)
	if !ok {
		return nil, errors.New("no executor found")
	}

	return f(ctx, req)
}

// Serve accepts incoming connections for handling.
func (s *Server) Serve() {
	rpc.RegisterBlobServer(s.gs, s)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port))
	if err != nil {
		s.cfg.Logger.Error().Err(err).Uint64("port", s.cfg.Port).Msg("creating grpc listener")
		return
	}

	err = s.gs.Serve(listener)
	if err != nil {
		s.cfg.Logger.Error().Err(err).Uint64("port", s.cfg.Port).Msg("serving grpc requests")
		return
	}
}

// Stop terminates the server.
func (s *Server) Stop() {
	s.gs.Stop()
}
