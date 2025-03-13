package blobrpc

import (
	"context"
	"fmt"

	"github.com/dnldd/blobrpc/rpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// ClientConfig represents the configuration of the blobrpc client.
type ClientConfig struct {
	// The server endpoint.
	Endpoint string
	// The grpc client configuration options.
	Options []grpc.DialOption
	// The service logger.
	Logger *zerolog.Logger
}

// Client represents a blobrpc client.
type Client struct {
	cfg *ClientConfig
	gc  *grpc.ClientConn
	cl  rpc.BlobClient
}

// NewClient initializes a new client.
func NewClient(cfg *ClientConfig) (*Client, error) {
	conn, err := grpc.NewClient(cfg.Endpoint, cfg.Options...)
	if err != nil {
		return nil, fmt.Errorf("Creating grpc client: %w", err)
	}

	cl := rpc.NewBlobClient(conn)

	return &Client{
		cfg: cfg,
		gc:  conn,
		cl:  cl,
	}, nil
}

// Send relays the provided request for processing.
func (c *Client) Send(ctx context.Context, in *rpc.Request, opts ...grpc.CallOption) (*rpc.Response, error) {
	resp, err := c.cl.Send(ctx, in, opts...)
	if err != nil {
		c.cfg.Logger.Error().Err(err).Str("route", in.Route).Uint32("version", in.Version).
			Str("marshal", in.Marshal_).Msg("sending request")
		return nil, err
	}

	return resp, nil
}

// Stop terminates the client.
func (c *Client) Stop() error {
	return c.gc.Close()
}
