package app

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"pi.com/lb/common"
	"pi.com/lb/model"
	"sync"
	"time"
)

type (
	lbApp struct {
		servers      []*serverApp
		currServerId int // currently for round-robin, will support more in future
		mux          *sync.Mutex
		healthChecks *model.Health
		logger       *logrus.Logger
	}

	serverApp struct {
		url          *url.URL
		isHealthy    bool
		reverseProxy *httputil.ReverseProxy
		mux          *sync.Mutex
	}
)

func NewLBApp(cfg *model.LoadBalancerConfig, logger *logrus.Logger) error {
	err := common.OnStartUpValidation(cfg)
	if err != nil {
		return fmt.Errorf("failed to start load balancer: %s", err)
	}

	app := &lbApp{
		servers:      make([]*serverApp, 0),
		mux:          new(sync.Mutex),
		currServerId: 0,
		healthChecks: cfg.HealthCheck,
		logger:       logger,
	}

	go app.checkIfServerHealthy()

	/*
		for _, pathRoutes := range cfg.PathRoutes {
			for _, server := range pathRoutes.Servers {
				parseUrl, err := url.Parse(server)
				if err != nil {
					return err
				}

				serverApp := &serverApp{
					mux:          new(sync.Mutex),
					url:          parseUrl,
					reverseProxy: httputil.NewSingleHostReverseProxy(parseUrl),
				}

				app.servers = append(app.servers, serverApp)
			}
		}

	*/

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Listen))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer listener.Close()

	err = http.Serve(listener, app)
	if err != nil {
		return err
	}
	return nil
}

func (lbApp *lbApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lbApp.mux.Lock()
	defer lbApp.mux.Unlock()

	server := &serverApp{}
	idx := 0

	for _ = range len(lbApp.servers) {
		idx = lbApp.currServerId % len(lbApp.servers)
		nextServers := lbApp.servers[idx]
		lbApp.currServerId++

		nextServers.mux.Lock()
		isHealthy := nextServers.isHealthy
		nextServers.mux.Unlock()
		if isHealthy {
			server = nextServers
			break
		}
	}

	if server != nil && server.url != nil {
		lbApp.logger.Debugf("%d: %s requesting endpoint", idx, server.url.String())
		server.reverseProxy.ServeHTTP(w, r)
	} else {
		lbApp.logger.Debugf("no endpoints available")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("no endpoint available"))
	}

}

func (lbApp *lbApp) checkIfServerHealthy() {
	ticker := time.NewTicker(time.Duration(lbApp.healthChecks.IntervalInSeconds) * time.Second)

	for _ = range ticker.C {
		for _, server := range lbApp.servers {
			go func() {
				healthCheckPath, err := url.JoinPath(server.url.String(), lbApp.healthChecks.Endpoint)
				if err != nil {
					lbApp.logger.Debugf("failed to join url path due to err: %s", err)
					return
				}
				response, err := http.Get(healthCheckPath)
				if err != nil {
					lbApp.logger.Debugf("failed to connect to health check due to err: %s", err)
					server.mux.Lock()
					server.isHealthy = false
					server.mux.Unlock()
					return
				}
				defer response.Body.Close()
				server.mux.Lock()

				if response.StatusCode != 200 {
					server.isHealthy = false
				} else {
					server.isHealthy = true
				}
				server.mux.Unlock()
			}()
		}
	}

}
