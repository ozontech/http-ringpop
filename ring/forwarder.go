package ring

import (
	"github.com/uber-common/bark"
	"github.com/uber/ringpop-go"
	"github.com/uber/tchannel-go"
)

// Forwarder is a request forwarder used to transfer request between nodes in hashring
type Forwarder interface {
	Forward(node, key string, request []byte) ([]byte, error)
}

// NewForwarder returns new request forwarder
func NewForwarder(rp *ringpop.Ringpop, l bark.Logger) *RequestForwarder {
	return &RequestForwarder{
		channelName: channelName,
		endpoint:    endpoint,
		ringpop:     rp,
		logger:      l,
	}
}

// RequestForwarder is a simple request forwarder that transfers request between nodes using ringpop
type RequestForwarder struct {
	channelName string
	endpoint    string

	ringpop *ringpop.Ringpop
	logger  bark.Logger
}

func (f *RequestForwarder) Forward(node, key string, request []byte) ([]byte, error) {
	f.logger.Infof(
		"Forwarding request to node: %s, key: %s, channel: %s, endpoint: %s",
		node, key, f.channelName, f.endpoint,
	)
	return f.ringpop.Forward(node, []string{key}, request, f.channelName, f.endpoint, tchannel.HTTP, nil)
}
