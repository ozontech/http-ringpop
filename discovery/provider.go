package discovery

import (
	"errors"

	"github.com/uber-common/bark"
	"github.com/uber/ringpop-go/discovery"
)

// providerBuilder is a simple builder for discovery provider
type providerBuilder struct {
	jsonFilePath string

	dnsHost     string
	dnsHostPort int

	logger bark.Logger
}

// NewProvider returns new discivery provider based on given arguments
func NewProviderBuilder(logger bark.Logger) *providerBuilder {
	return &providerBuilder{
		logger: logger,
	}
}

func (b *providerBuilder) WithJSONFileDiscovery(jsonFilePath string) *providerBuilder {
	b.jsonFilePath = jsonFilePath
	return b
}

func (b *providerBuilder) WithDNSDiscovery(dnsHost string, dnsHostPort int) *providerBuilder {
	b.dnsHost = dnsHost
	b.dnsHostPort = dnsHostPort
	return b
}

func (b *providerBuilder) Build() (discovery.DiscoverProvider, error) {
	if b.jsonFilePath != "" {
		return newJSONFileDiscovery(b.jsonFilePath), nil
	}

	if b.dnsHost != "" && b.dnsHostPort != 0 {
		return newDNSProvider(b.dnsHost, b.dnsHostPort, b.logger), nil
	}

	return nil, errors.New("Not enough arguments for building discovery provider")
}
