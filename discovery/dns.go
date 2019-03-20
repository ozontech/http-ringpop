package discovery

import (
	"fmt"
	"net"
	"strings"

	"github.com/uber-common/bark"
	"github.com/uber/ringpop-go/discovery"
)

// DNSProvider returns a list of host hosts
// Compatible with github.com/uber/ringpop-go/discovery.DiscoveryProvider interface
type DNSProvider struct {
	host   string
	port   int
	logger bark.Logger
}

// newDNSProvider returns new providers that discovers host endpoints
// by hostname using DNS records
func newDNSProvider(host string, port int, logger bark.Logger) discovery.DiscoverProvider {
	provider := &DNSProvider{
		host:   host,
		port:   port,
		logger: logger,
	}
	return provider
}

func (k *DNSProvider) Hosts() ([]string, error) {
	k.logger.Infof("Discovering hosts from DNS by hostname: %s...", k.host)

	addrs, err := net.LookupHost(k.host)
	if err != nil {
		if _, ok := err.(*net.DNSError); ok {
			return []string{}, nil
		}

		return nil, err
	}

	for i := range addrs {
		addrs[i] = fmt.Sprintf("%s:%d", addrs[i], k.port)
	}

	k.logger.Infof("Discovered endpoints: %s", strings.Join(addrs, ", "))

	return addrs, nil
}
