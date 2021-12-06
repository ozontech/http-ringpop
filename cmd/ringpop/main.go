package main

import (
	"flag"
	"net"
	"net/http"
	"os"

	"github.com/ozontech/http-ringpop/backend"
	"github.com/ozontech/http-ringpop/discovery"
	ringhttp "github.com/ozontech/http-ringpop/http"
	"github.com/ozontech/http-ringpop/pkg/metrics"
	"github.com/ozontech/http-ringpop/ring"

	"github.com/sirupsen/logrus"
	"github.com/uber-common/bark"
)

var (
	httpListenOn    = flag.String("listen.http", ":3000", "hostPort to listen calls from incoming http requests")
	backendURL      = flag.String("backend.url", "http://127.0.0.1:4000/", "URL of your http backend")
	ringpopListenOn = flag.String("listen.ringpop", ":5000", "hostPort to listen gossip requests inside hashring")
	debugListenOn   = flag.String("listen.debug", ":6000", "hostPort to listen calls from incoming debug http requests (metrics, etc.)")
	logLevel        = flag.Uint("log.level", 4, "Log level, default - INFO (4)")

	discoveryJSONFile = flag.String("discovery.json.file", "", "Discovery hosts from static file")

	discoveryDNSHost     = flag.String("discovery.dns.host", "", "Discovery hosts from DNS by hostname")
	discoveryDNSHostPort = flag.Int("discovery.dns.port", 0, "Ringpop port that will be added to discovered hosts from DNS")

	// ringpopPeerIP is required to identify current instance in the ring.
	// For example, when instances will be discovered from DNS records, it will be something like this:
	//	10.27.27.42:5000
	//	10.27.35.133:5000
	// Current IP could be detected correctly only in particular cases. So, if current IP will be
	// detected automatically as 127.0.0.1 this node will try to join to itself.
	ringpopPeerIP = os.Getenv("RINGPOP_PEER_IP")
)

func main() {
	flag.Parse()

	l := logrus.StandardLogger()
	l.Level = logrus.Level(*logLevel)
	logger := bark.NewLoggerFromLogrus(logrus.StandardLogger())

	backendProxy, err := backend.New(*backendURL, logger)
	if err != nil {
		logger.Fatalf("unable to create backend reverse backendProxy: %v", err)
	}

	ch, err := ring.NewChannel()
	if err != nil {
		logger.Fatalf("unable to create channel: %v", err)
	}

	ringpopPeerIP, ringpopPeerPort, err := getRingpopPeerHostPort(ringpopPeerIP, *ringpopListenOn)
	if err != nil {
		logger.Fatalf("unable to resolve ringpopPeerIP or ringpopPeerPort: %v", err)
	}

	ring.SetLogger(logger)
	rp, err := ring.NewRingpop(ch, ringpopPeerIP, ringpopPeerPort, logger)
	if err != nil {
		logger.Fatalf("unable to create Ringpop: %v", err)
	}

	logger.Info("Running ringpop server...")
	ringpopServer := ring.NewServer(ch, backendProxy, logger)

	if err := ringpopServer.ListenAndServe(*ringpopListenOn); err != nil {
		logger.Fatalf("unable to listen on given addr: %v", err)
	}
	logger.Info("...OK")

	discoveryBuilder := discovery.NewProviderBuilder(logger)
	if *discoveryJSONFile != "" {
		discoveryBuilder.WithJSONFileDiscovery(*discoveryJSONFile)
	}
	if *discoveryDNSHost != "" && *discoveryDNSHostPort != 0 {
		discoveryBuilder.WithDNSDiscovery(*discoveryDNSHost, *discoveryDNSHostPort)
	}

	discoveryProvider, err := discoveryBuilder.Build()
	if err != nil {
		logger.Fatalf("unable to create discovery provider: %v", err)
	}

	go func() {
		requestForwarder := ring.NewForwarder(rp, logger)

		logger.Infof("Running HTTP reverse proxy server on %s for backend %s...", *httpListenOn, *backendURL)
		// Transparent front HTTP server
		httpServer := ringhttp.NewServer(rp, requestForwarder, backendProxy, logger)

		http.HandleFunc("/", httpServer.Handle)
		if err := http.ListenAndServe(*httpListenOn, nil); err != nil {
			logger.Fatalf("unable to listen on %s: %s", *httpListenOn, err)
		}

		logger.Info("...OK")
	}()

	go func() {
		debugSrv := http.NewServeMux()
		debugSrv.Handle(metrics.MetricsPath, metrics.Handler())

		logger.Infof("Running debug HTTP server on %s...", *debugListenOn)

		if err := http.ListenAndServe(*debugListenOn, debugSrv); err != nil {
			logger.Fatalf("unable to listen on %s: %s", *debugListenOn, err)
		}
	}()

	logger.Infof("Bootstrapping ringpop on %s...", *ringpopListenOn)
	if err := ring.BootstrapRingpop(rp, discoveryProvider); err != nil {
		logger.Fatalf("ringpop bootstrap failed: %v", err)
	}
	logger.Info("...OK")

	select {}
}

func getRingpopPeerHostPort(ringpopPeerIP, pingpopListenOn string) (ip, port string, err error) {
	if ringpopPeerIP != "" {
		ip = ringpopPeerIP
	}

	host, port, err := net.SplitHostPort(pingpopListenOn)
	if err != nil {
		return
	}

	if host == "" {
		host = "127.0.0.1"
	}

	if ip == "" {
		ip = host
	}

	return
}
