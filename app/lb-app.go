package app

import (
	"errors"
	"net/http/httputil"
	"net/url"
	"pi.com/lb/model"
	"sync"
)

type (
	lbApp struct {
		servers      []*serverApp
		currServerId int // currently for round-robin, will support more in future
		mux          *sync.Mutex
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

	for _, server := range cfg.ServerList {
		parseUrl, err := url.Parse(server)
		if err != nil {
			return err
		}

		serverApp := &serverApp{
			mux:          new(sync.Mutex),
			url:          parseUrl,
			isHealthy:    true,
			reverseProxy: httputil.NewSingleHostReverseProxy(parseUrl),
		}

		lbApp.servers = append(lbApp.servers, serverApp)
	}

	return nil
}
