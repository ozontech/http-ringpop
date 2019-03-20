package discovery

import (
	"github.com/uber/ringpop-go/discovery/jsonfile"
)

// newJSONFileDiscovery returns static discovery provider
// with hosts list in JSON file.
// Compatible with github.com/uber/ringpop-go/discovery.DiscoveryProvider interface
//
// JSON file example:
// ["127.0.0.1:3000", "127.0.0.1:3001"]
func newJSONFileDiscovery(filePath string) (*jsonfile.HostList) {
	return jsonfile.New(filePath)
}
