package blobrpc

import (
	"context"
	"testing"

	"github.com/dnldd/blobrpc/rpc"
	"github.com/peterldowns/testy/assert"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.With().Caller().Logger()
	handler := NewTestHandler()

	// Create the blobrpc server.
	sCfg := &ServerConfig{
		Port:    3556,
		Options: []grpc.ServerOption{},
		Handler: handler,
		Logger:  &logger,
	}

	done := make(chan bool)
	bs := NewServer(sCfg)
	go func(s *Server) {
		bs.Serve()
		close(done)
	}(bs)

	// Create the blobrpc client.
	cCfg := &ClientConfig{
		Endpoint: "0.0.0.0:3556",
		Options: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
		Logger: &logger,
	}
	cc, err := NewClient(cCfg)
	assert.NoError(t, err)

	// Ensure a valid request gets processed as expected.
	req := &rpc.Request{
		Route:    "echo",
		Version:  1,
		Marshal_: "plain",
		Payload:  []byte("hello world"),
	}

	resp, err := cc.Send(ctx, req, []grpc.CallOption{}...)
	assert.NoError(t, err)
	assert.Equal(t, string(resp.Payload), "hello world")

	// Ensure an invalid route generates an error.
	req = &rpc.Request{
		Route:    "invalid_route",
		Version:  1,
		Marshal_: "plain",
		Payload:  []byte("hello world"),
	}

	_, err = cc.Send(ctx, req, []grpc.CallOption{}...)
	assert.Error(t, err)

	// Ensure requests can have an empty payload.
	req = &rpc.Request{
		Route:    "echo",
		Version:  1,
		Marshal_: "plain",
		Payload:  []byte(""),
	}

	resp, err = cc.Send(ctx, req, []grpc.CallOption{}...)
	assert.NoError(t, err)
	assert.Equal(t, string(resp.Payload), "")

	// Ensure an unknown request version generates an error.
	req = &rpc.Request{
		Route:    "echo",
		Version:  2,
		Marshal_: "plain",
		Payload:  []byte(""),
	}

	resp, err = cc.Send(ctx, req, []grpc.CallOption{}...)
	assert.Error(t, err)

	// Ensure the server can be stopped.
	bs.Stop()
	<-done
	cancel()
}
