package splunktracing

import (
	"context"
	"io"
	"net/http"

)

var accessTokenHeader = http.CanonicalHeaderKey("Authorization")

// Connection describes a closable connection. Exposed for testing.
type Connection interface {
	io.Closer
}

// ConnectorFactory is for testing purposes.
type ConnectorFactory func() (interface{}, Connection, error)

// collectorResponse encapsulates internal grpc/http responses.
type collectorResponse interface {
	GetErrors() []string
	Disable() bool
	DevMode() bool
}

type reportRequest struct {
	httpRequest  *http.Request
}

// collectorClient encapsulates internal grpc/http transports.
type collectorClient interface {
	Report(context.Context, reportRequest) (collectorResponse, error)
	Translate(context.Context, *reportBuffer) (reportRequest, error)
	ConnectClient() (Connection, error)
	ShouldReconnect() bool
}

func newCollectorClient(opts Options, reporterID uint64, attributes map[string]string) (collectorClient, error) {

	// No transport specified, defaulting to HTTP
	return newHTTPCollectorClient(opts, reporterID, attributes)
}
