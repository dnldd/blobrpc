# BlobRPC

BlobRPC is a gRPC-based inter-service communication protocol that uses generic, self-describing request and response types. Opting for generic, self-describing request and responses allows relaying blobs as payloads over a standard gRPC connection. 

The request and response types are described as self-describing since they indicate key attributes of 
the blob enveloped. The types describe the handler the blob should be handled with via the `route` and `version` fields. The types also describe the serialization of the blob via the `marshal` field.

## Usage

### RequestHandler

The BlobRPC server requires a `RequestHandler` in order to process incoming messages.

```go
// RequestHandler defines the requirements of a blobrpc request handler.
type RequestHandler interface {
	// ID generates the executor id from the provided route and version.
	ID(route string, version uint32) string
	// FetchExecutor fetches the executor for the provided id.
	FetchExecutor(id string) (Executor, bool)
}
```

Here is a simple handler implementation:

```go
// Handler represents a request handler.
type Handler struct {
	executors map[string]Executor
}

// NewHandler initializes a new handler.
func NewHandler() *Handler {
	handler := Handler{executors: make(map[string]Executor)}

	// Map the handlers to their respective ids.
	handler.executors["echo.v1"] = handler.Echo

	return &handler
}

// ID generates the executor id from the provided route and version.
func (h *Handler) ID(route string, version uint32) string {
	return fmt.Sprintf("%s.v%d", route, version)
}

// FetchExecutor fetches the executor for the provided id.
func (h *Handler) FetchExecutor(id string) (Executor, bool) {
	executor, ok := h.executors[id]
	return executor, ok
}

// Echo responds with the request blob as the response.
func (h *TestHandler) Echo(_ context.Context, req *rpc.Request) (*rpc.Response, error) {
	return &rpc.Response{
		Route:    req.Route,
		Version:  req.Version,
		Marshal_: req.Marshal_,
		Payload:  req.Payload,
		Error:    "",
	}, nil
}
```

### Server

To create a BlobRPC server with the handler:

```go
    logger := log.With().Caller().Logger()
    handler := blobrpc.NewTestHandler()

    cfg := &blobrpc.ServerConfig{
        Port:    3555,
        Options: []grpc.ServerOption{},
        Handler: handler,
        Logger:  &logger,
    }
    server := blobrpc.NewServer(cfg)
    go server.Serve()
```

### Client

To create a BlobRPC client:

```go
    logger := log.With().Caller().Logger()

    cfg := &blobrpc.ClientConfig{
        Endpoint: "0.0.0.0:3555",
        Options: []grpc.DialOption{
            grpc.WithTransportCredentials(insecure.NewCredentials()),
        },
        Logger: &logger,
    }
    client, err := blobrpc.NewClient(cfg)
    if err != nil {
        logger.Fatal().Err(err).Msg("creating client")
    }

    req := &rpc.Request{
        Route:    "echo",
        Version:  1,
        Marshal_: "plain",
        Payload:  []byte("hello world"),
    }

    resp, err := client.Send(context.Background(), req)
    if err != nil {
        logger.Error().Err(err).Msg("sending request")
    } else {
        logger.Info().Str("response", string(resp.Payload)).Msg("received response")
    }
```
