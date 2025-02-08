package app

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"pi.com/lb/model"
	"sync"
	"time"
)

type (
	lbApp struct {
		servers      []*serverApp
		currServerId int // currently for round-robin, will support more in future
		mux          *sync.Mutex
		healthChecks *model.HealthCheckCfg
	}

	serverApp struct {
		url          *url.URL
		isHealthy    bool
		reverseProxy *httputil.ReverseProxy
		mux          *sync.Mutex
	}
)

func NewLBApp(cfg *model.Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	app := &lbApp{
		servers:      make([]*serverApp, 0),
		mux:          new(sync.Mutex),
		currServerId: 0,
		healthChecks: cfg.HealthCheckCfg,
	}

	go app.checkIfServerHealthy()

	for _, server := range cfg.ServerList {
		parseUrl, err := url.Parse(server)
		if err != nil {
			return err
		}

		serverApp := &serverApp{
			mux:          new(sync.Mutex),
			url:          parseUrl,
			reverseProxy: httputil.NewSingleHostReverseProxy(parseUrl),
		}

		/* FIXME: check if required
		serverApp.reverseProxy.Director = func(req *http.Request) {
			req.URL.Scheme = parseUrl.Scheme
			req.URL.Host = parseUrl.Host
		}

		*/

		app.servers = append(app.servers, serverApp)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
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
	server := lbApp.servers[lbApp.currServerId]
	server.reverseProxy.ServeHTTP(w, r)

}

func (lbApp *lbApp) checkIfServerHealthy() {
	ticker := time.NewTicker(time.Duration(lbApp.healthChecks.IntervalInSeconds) * time.Second)

	for _ = range ticker.C {
		for _, server := range lbApp.servers {
			go func() {
				healthCheckPath, err := url.JoinPath(server.url.String(), lbApp.healthChecks.Endpoint)
				if err != nil {
					return
				}
				response, err := http.Get(healthCheckPath)
				if err != nil {
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
