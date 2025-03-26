package blobrpc

import (
	"context"
	"fmt"
	"testing"

	"github.com/dnldd/blobrpc/rpc"
	"github.com/peterldowns/testy/assert"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Test handler represents a blobrpc handler, simplified for testing.
type TestHandler struct {
	executors map[string]Executor
}

func NewTestHandler() *TestHandler {
	handler := TestHandler{
		executors: make(map[string]Executor),
	}

	handler.executors["echo.v1"] = handler.Echo

	return &handler
}

// ID generates the executor id from the provided route and version.
func (h *TestHandler) ID(route string, version uint32) string {
	return fmt.Sprintf("%s.v%d", route, version)
}

// FetchExecutor fetches the executor for the provided id.
func (h *TestHandler) FetchExecutor(id string) (Executor, bool) {
	executor, ok := h.executors[id]
	return executor, ok
}

func (h *TestHandler) Echo(_ context.Context, req *rpc.Request) (*rpc.Response, error) {
	return &rpc.Response{
		Route:    req.Route,
		Version:  req.Version,
		Marshal_: req.Marshal_,
		Payload:  req.Payload,
		Error:    "",
	}, nil
}

func TestRequestResponse(t *testing.T) {
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

	resp, err := cc.Send(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, string(resp.Payload), "hello world")

	// Ensure an invalid route generates an error.
	req = &rpc.Request{
		Route:    "invalid_route",
		Version:  1,
		Marshal_: "plain",
		Payload:  []byte("hello world"),
	}

	_, err = cc.Send(ctx, req)
	assert.Error(t, err)

	// Ensure requests can have an empty payload.
	req = &rpc.Request{
		Route:    "echo",
		Version:  1,
		Marshal_: "plain",
		Payload:  []byte(""),
	}

	resp, err = cc.Send(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, string(resp.Payload), "")

	// Ensure an unknown request version generates an error.
	req = &rpc.Request{
		Route:    "echo",
		Version:  2,
		Marshal_: "plain",
		Payload:  []byte(""),
	}

	resp, err = cc.Send(ctx, req)
	assert.Error(t, err)

	// Ensure the server can be terminated.
	bs.Stop()

	// Ensure sending to terminated server errors.
	resp, err = cc.Send(ctx, req)
	assert.Error(t, err)

	<-done
	cancel()
}
