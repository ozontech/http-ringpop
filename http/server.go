package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ozontech/http-ringpop/pkg/metrics"
	"github.com/ozontech/http-ringpop/ring"

	"github.com/uber-common/bark"
	"github.com/uber/ringpop-go"
)

const (
	headerProxy             = "X-Proxy"
	headerRingpopReceivedBy = "X-Ringpop-Received-By"
	headerRingpopHandledBy  = "X-Ringpop-Handled-By"
)

var (
	metricHTTPRequestsTotal               = metrics.MustRegisterCounter("http_requests_total", "Total number of received HTTP requests")
	metricRequestsForwardedToBackendTotal = metrics.MustRegisterCounter("requests_forwarded_to_backend_total", "Total number of requests forwarded to HTTP backend")
	metricRequestsForwardedToRingpopTotal = metrics.MustRegisterCounter("requests_forwarded_to_ringpop_total", "Total number of requests forwarded to ringpop")
)

// NewServer returns new HTTPServer
func NewServer(rp *ringpop.Ringpop, f ring.Forwarder, backend http.Handler, l bark.Logger) *HTTPServer {
	return &HTTPServer{
		ringpop:          rp,
		requestForwarder: f,
		backend:          backend,
		logger:           l,
	}
}

// HTTPServer serves all incoming HTTP requests
type HTTPServer struct {
	ringpop          *ringpop.Ringpop
	requestForwarder ring.Forwarder
	backend          http.Handler
	logger           bark.Logger
}

// Handle
func (srv *HTTPServer) Handle(w http.ResponseWriter, r *http.Request) {
	metricHTTPRequestsTotal.Inc()

	key := ring.RequestToKey(r)
	srv.logger.Infof("Got request. Key: %s", key)

	dstNode, err := ring.ResolveDestinationNode(srv.ringpop, key)
	if err != nil {
		srv.logger.Errorf("Can't resolve dst node: %v", err)
		fmt.Fprintf(w, "Can't resolve dst node: %s", err)
		return
	}

	address, err := srv.ringpop.WhoAmI()
	if err != nil {
		srv.logger.Errorf("Can't resolve who am I: %v", err)
		fmt.Fprintf(w, "Can't resolve who am I: %s", err)
		return
	}

	w.Header().Set(headerRingpopReceivedBy, address)
	r.Header.Set(headerRingpopReceivedBy, address) // Just to know on dst node who was first receiver

	shouldHandle := address == dstNode

	if shouldHandle {
		srv.logger.Info("Request will be handled current node, proxying request to backend...")

		r.Header.Set(headerProxy, address)
		w.Header().Set(headerRingpopHandledBy, address)

		// ServeHTTP request on this instance
		srv.backend.ServeHTTP(w, r)

		metricRequestsForwardedToBackendTotal.Inc()

		return
	}

	srv.logger.Infof("Request will be handled on another node: %v", dstNode)

	// Forward request to responsible host
	srv.forwardRequestToDstNode(dstNode, key, w, r)
}

func (srv *HTTPServer) forwardRequestToDstNode(dstNode, key string, w http.ResponseWriter, r *http.Request) {
	// Override request host (it doesn't affect anything, just for consistency)
	r.Host = dstNode

	requestBytes, err := httpRequestToBytes(r)
	if err != nil {
		fmt.Fprintf(w, "Unable to write incoming request to buffer: %v", err)
		srv.logger.Errorf("Unable to write incoming request to buffer: %v", err)
		return
	}

	rawResponse, err := srv.requestForwarder.Forward(dstNode, key, requestBytes)
	if err != nil {
		fmt.Fprintf(w, "Unable to forward request: %v", err)
		srv.logger.Errorf("Unable to forward request: %v", err)
		return
	}

	metricRequestsForwardedToRingpopTotal.Inc()

	w.Header().Set(headerRingpopHandledBy, dstNode)

	if err := copyHTTPResponseFromRaw(w, r, rawResponse); err != nil {
		fmt.Fprintf(w, "Unable to copy response from raw: %v", err)
		srv.logger.Errorf("Unable to copy response from raw: %v", err)
		return
	}
}

func httpRequestToBytes(r *http.Request) ([]byte, error) {
	request := &bytes.Buffer{}

	if err := r.Write(request); err != nil {
		return nil, err
	}

	return request.Bytes(), nil
}

// copyHTTPResponseFromRaw copies data from rawResponse to responseWriter
func copyHTTPResponseFromRaw(w http.ResponseWriter, r *http.Request, rawResponse []byte) error {
	buf := bufio.NewReader(bytes.NewReader(rawResponse))
	resp, err := http.ReadResponse(buf, r)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	for k := range resp.Header {
		w.Header().Set(k, resp.Header.Get(k))
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	return nil
}
