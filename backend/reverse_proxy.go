package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/uber-common/bark"
)

// BackendReverseProxy is a wrapper around standard http proxy
type BackendReverseProxy struct {
	proxy  *httputil.ReverseProxy
	target *url.URL
	logger bark.Logger
}

// New returns new reverse proxy for given backend
func New(target string, logger bark.Logger) (*BackendReverseProxy, error) {
	uri, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	return &BackendReverseProxy{
		proxy:  httputil.NewSingleHostReverseProxy(uri),
		target: uri,
		logger: logger,
	}, nil
}

func (b *BackendReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.logger.Infof("Proxying request to HTTP backend: %s", b.target.String())

	req, _ := httputil.DumpRequest(r, true)
	b.logger.Debugf("Request to by proxied:\n------------\n%s\n------------", string(req))

	b.proxy.ServeHTTP(w, r)
}
