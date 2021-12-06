package ring

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/ozontech/http-ringpop/pkg/metrics"

	"github.com/uber-common/bark"
	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/raw"
)

const (
	appName     = "distributed-hashring-app"
	channelName = "ringpop-channel"
	endpoint    = "/request-exchange"
)

var (
	metricRingpopRequestsTotal = metrics.MustRegisterCounter("ringpop_requests_total", "Total number of received ringpop requests")
)

// Server serves ringpop (gossip) requests inside ring
type Server struct {
	channel  *tchannel.Channel
	endpoint string

	backend http.Handler

	logger bark.Logger
}

// NewServer returns new ringpop server and registers its handlers
func NewServer(ch *tchannel.Channel, backend http.Handler, logger bark.Logger) *Server {
	srv := &Server{
		channel:  ch,
		endpoint: endpoint,
		backend:  backend,
		logger:   logger,
	}

	srv.registerHandlers()

	return srv
}

// ListenAndServe starts listening on TChannel
func (srv *Server) ListenAndServe(hostPort string) error {
	return srv.channel.ListenAndServe(hostPort)
}

func (srv *Server) registerHandlers() error {
	handler := raw.Wrap(ringpopRequestHandler{
		backend: srv.backend,
		logger:  srv.logger,
	})
	srv.channel.Register(handler, srv.endpoint)

	return nil
}

// ringpopRequestHandler is a handle for ringpop requests
type ringpopRequestHandler struct {
	backend http.Handler
	logger  bark.Logger
}

// Handle is a ringpop request handler (it works only for gossip communication inside ring)
//
// Its expected that this func will be called only on responsible server:
// Request forwarding (if it was needed) was already done in HTTPServer.forwardRequestToDstNode()
// So we assume that no additional forwarding required
// There are one possible problem - if size of hashring was changed when request was in progress -
// this node could be already not responsible of this request. But its minor thing and we're skipping it at the moment.
func (h ringpopRequestHandler) Handle(ctx context.Context, args *raw.Args) (*raw.Res, error) {
	metricRingpopRequestsTotal.Inc()

	h.logger.Infof("Got request, caller: %s, method: %s, format: %s", args.Caller, args.Method, args.Format)
	h.logger.Debugf("Arg2:\n---\n%s---", string(args.Arg2))
	h.logger.Debugf("Arg3:\n---\n%s---", string(args.Arg3))

	requestReader := bufio.NewReader(bytes.NewReader(args.Arg3))
	request, err := http.ReadRequest(requestReader)
	if err != nil {
		h.logger.Errorf("Error on reading request from raw data: %v", err)
	}

	respWriter := NewResponseWriter()

	// Serve request on HTTP backend
	h.backend.ServeHTTP(respWriter, request)

	rawResponse := []byte{}
	buffer := bytes.NewBuffer(rawResponse)

	if err := respWriter.Response().Write(buffer); err != nil {
		h.logger.Errorf("Error on writing buffer: %v", err)
	}

	b, _ := ioutil.ReadAll(buffer)
	h.logger.Debugf("Response written:\n------------\n%s\n------------", string(b))

	return &raw.Res{
		Arg2: args.Arg2,
		Arg3: b,
	}, nil
}

func (h ringpopRequestHandler) OnError(ctx context.Context, err error) {
	h.logger.Errorf("OnError: %v", err)
}
