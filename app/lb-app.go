package app

import (
	"fmt"
	"github.com/armon/go-radix"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"pi.com/lb/common"
	"pi.com/lb/constants"
	"pi.com/lb/model"
	"sync"
	"time"
)

type (
	lbApp struct {
		strategy model.LoadBalanceStrategy
		//upstream   map[string][]*serverInfo
		upstream map[string]*upstreamInfo

		pathRoutes *radix.Tree

		healthChecks *model.Health
		logger       *logrus.Logger
		//*metaInfo
	}

	upstreamInfo struct {
		serverInfo []*serverInfo
		*metaInfo
	}

	serverInfo struct {
		url          *url.URL
		isHealthy    bool
		reverseProxy *httputil.ReverseProxy
		mux          *sync.Mutex
	}

	pathRouteInfo struct {
		proxyPass string
	}

	metaInfo struct {
		currServerId   int
		mux            *sync.Mutex
		currConnection int
	}
)

func NewLBApp(cfg *model.LoadBalancerConfig, logger *logrus.Logger) error {
	err := common.OnStartUpValidation(cfg)
	if err != nil {
		return fmt.Errorf("failed to start load balancer: %s", err)
	}

	app := &lbApp{
		healthChecks: cfg.HealthCheck,
		logger:       logger,
		upstream:     make(map[string]*upstreamInfo),
		pathRoutes:   radix.New(),
		strategy:     cfg.Strategy,
	}

	err = app.mapConfig(cfg)
	if err != nil {
		return err
	}

	app.strategySetup()

	go app.checkIfServerHealthy()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Listen))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer listener.Close()

	logger.Debugf("started load balancer on port %d", cfg.Listen)
	err = http.Serve(listener, app)
	if err != nil {
		return err
	}
	return nil
}

func (lbApp *lbApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//FIXME: add validation for server address

	switch lbApp.strategy {
	case model.ROUND_ROBIN:
		lbApp.roundRobinImpl(w, r)
	}

}

func (lbApp *lbApp) checkIfServerHealthy() {
	interval := constants.DEFAULT_HEALTH_CHECK_INTERVAL
	healthCheckPath := ""

	if lbApp.healthChecks != nil {

		if lbApp.healthChecks.IntervalInSeconds > 0 {
			interval = lbApp.healthChecks.IntervalInSeconds
		}

		if common.IsStringNonEmpty(lbApp.healthChecks.Endpoint) {
			healthCheckPath = lbApp.healthChecks.Endpoint
		}

	}

	timeTicker := time.NewTicker(time.Duration(interval) * time.Second)

	for _ = range timeTicker.C {
		go func() {
			for name, upstream := range lbApp.upstream {

				for _, server := range upstream.serverInfo {
					err := common.ServerURLValidation(server.url.String(), healthCheckPath, 0)

					server.mux.Lock()
					if err != nil {
						lbApp.logger.Errorf("upstream %s : server: %s unhealthy due to err: %s", name, server.url.String(), err)
						server.isHealthy = false
					} else {
						lbApp.logger.Tracef("upstream %s : server: %s healthy", name, server.url.String())
						server.isHealthy = true
					}
					server.mux.Unlock()
				}
			}
		}()

	}

}

func (lbApp *lbApp) mapConfig(cfg *model.LoadBalancerConfig) error {

	for name, upstream := range cfg.Upstream {
		lbApp.upstream[name] = &upstreamInfo{
			serverInfo: make([]*serverInfo, 0),
		}

		for _, server := range upstream {
			parseURL, err := url.Parse(server.URL)
			if err != nil {
				return err
			}

			info := &serverInfo{
				isHealthy:    true,
				reverseProxy: httputil.NewSingleHostReverseProxy(parseURL),
				mux:          &sync.Mutex{},
				url:          parseURL,
			}

			lbApp.upstream[name].serverInfo = append(lbApp.upstream[name].serverInfo, info)
		}
	}

	for _, location := range cfg.Location {
		lbApp.pathRoutes.Insert(location.Path, &pathRouteInfo{proxyPass: location.ProxyPass})
	}

	return nil
}

func (lbApp *lbApp) strategySetup() {
	switch lbApp.strategy {
	case model.ROUND_ROBIN:

		for _, upstream := range lbApp.upstream {
			upstream.metaInfo = &metaInfo{
				currServerId: 0,
				mux:          &sync.Mutex{},
			}
		}
	case model.LEAST_CONN:
		for _, upstream := range lbApp.upstream {
			upstream.metaInfo = &metaInfo{
				currConnection: 0,
				mux:            &sync.Mutex{},
			}
		}
	}
}

func (lbApp *lbApp) mostMatchingLocation(requestPath string, w http.ResponseWriter) (*url.URL, error) {

	_, bestMatch, found := lbApp.pathRoutes.LongestPrefix(requestPath)
	if !found {
		http.Error(w, "Service Not Found", http.StatusNotFound)
		return nil, fmt.Errorf("service Not Found")
	}

	route := bestMatch.(*pathRouteInfo)

	// Parse the proxy URL
	parseURL, err := url.Parse(route.proxyPass)
	if err != nil {
		http.Error(w, "Invalid Proxy URL", http.StatusInternalServerError)
		return nil, fmt.Errorf("invalid Proxy URL")
	}

	return parseURL, nil
}
