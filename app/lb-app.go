package app

import (
	"errors"
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

	lbApp := &lbApp{
		servers:      make([]*serverApp, 0),
		mux:          new(sync.Mutex),
		currServerId: 0,
	}

	go lbApp.checkIfServerHealthy()

	for _, server := range cfg.ServerList {
		parseUrl, err := url.Parse(server)
		if err != nil {
			return err
		}

		serverApp := &serverApp{
			mux: new(sync.Mutex),
			url: parseUrl,
			//isHealthy:    true,
			reverseProxy: httputil.NewSingleHostReverseProxy(parseUrl),
		}

		lbApp.servers = append(lbApp.servers, serverApp)
	}

	return nil
}

func (lbApp *lbApp) checkIfServerHealthy() {
	ticker := time.NewTicker(time.Duration(lbApp.healthChecks.IntervalInSeconds) * time.Second)

	for _ = range ticker.C {
		for _, server := range lbApp.servers {
			go func() {
				healthCheckPath, err := url.JoinPath(server.url.Host, lbApp.healthChecks.Endpoint)
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
