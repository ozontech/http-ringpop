package ring

import (
	"errors"
	"fmt"
	
	"github.com/uber-common/bark"
	"github.com/uber/ringpop-go"
	"github.com/uber/ringpop-go/discovery"
	"github.com/uber/ringpop-go/logging"
	"github.com/uber/ringpop-go/swim"
	"github.com/uber/tchannel-go"
)

var errorRingpopIsNotReady = errors.New("Ringpop is not ready")

// NewChannel returns new TChannel used for communication between nodes in ring
func NewChannel() (*tchannel.Channel, error) {
	return tchannel.NewChannel(channelName, nil)
}

// NewRingpop returns new ringpop
func NewRingpop(ch *tchannel.Channel, ringpopPeerIP, ringpopPeerPort string, logger bark.Logger) (*ringpop.Ringpop, error) {
	addr := fmt.Sprintf("%s:%s", ringpopPeerIP, ringpopPeerPort)

	return ringpop.New(
		appName,
		ringpop.Channel(ch),
		ringpop.Address(addr),
		ringpop.Logger(logger),
	)
}

// Bootstrap starts communication for this Ringpop instance.
// When Bootstrap is called, this Ringpop instance will attempt to contact
// other instances from the DiscoverProvider.
func BootstrapRingpop(rp *ringpop.Ringpop, provider discovery.DiscoverProvider) error {
	opts := &swim.BootstrapOptions{
		DiscoverProvider: provider,
	}

	_, err := rp.Bootstrap(opts)
	return err
}

// ResolveDestinationNode finds out responsible node from hashring by given key
func ResolveDestinationNode(rp *ringpop.Ringpop, key string) (string, error) {
	if !rp.Ready() {
		return "", errorRingpopIsNotReady
	}

	dest, err := rp.Lookup(key)
	if err != nil {
		return "", err
	}

	return dest, nil
}

// SetLogger sets default logger for ringpop
func SetLogger(logger bark.Logger) {
	logging.SetLogger(logger)
}
